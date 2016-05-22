/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/
package html

import (
	"bytes"
	"fmt"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/report"
	"hash/crc32"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func Generate(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	if err := generateHomePage(config, progressMonitor); err != nil {
		return err
	}
	if err := generateReports(config, progressMonitor); err != nil {
		return err
	}
	if err := generateTagPages(config); err != nil {
		return err
	}
	return generateProgressPage(config, progressMonitor)
}

func generateHomePage(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<meta charset="UTF-8">
		<title>Rulehunter</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Rulehunter</h1>
				Find simple rules in your data to meet your goals.

				<h2>Source</h2>
				Copyright (C) 2016 <a href="http://vlifesystems.com">vLife Systems Ltd</a>

				<p>This program is free software: you can redistribute it and/or modify
				it under the terms of the GNU Affero General Public License as published by
				the Free Software Foundation, either version 3 of the License, or
				(at your option) any later version.</p>

				<p>This program is distributed in the hope that it will be useful,
				but WITHOUT ANY WARRANTY; without even the implied warranty of
				MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
				GNU Affero General Public License for more details.</p>

				<p>You should have received a copy of the GNU Affero General Public License
				along with this program.  If not, see
				<a href="http://www.gnu.org/licenses/">http://www.gnu.org/licenses/</a>.</p>

				<p>The source is available on github: <a href="https://github.com/vlifesystems/rulehuntersrv">https://github.com/vlifesystems/rulehuntersrv</a>.</p>

			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplData struct {
		Html map[string]template.HTML
	}

	tplData := TplData{
		makeHtml("home"),
	}

	outputFilename := filepath.Join(config.WWWDir, "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateReports(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<meta charset="UTF-8">
		<title>Reports</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports</h1>

				<ul class="reports">
					{{range .Reports}}
						<li>
							<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
							Date: {{ .Stamp }}
							Tags:
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
							{{end}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplReport struct {
		Title    string
		Tags     map[string]string
		Stamp    string
		Filename string
	}

	type TplData struct {
		Reports []*TplReport
		Html    map[string]template.HTML
	}

	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	i := 0
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			reportFilename := makeReportFilename(report.Stamp, report.Title)
			if err = generateReport(report, reportFilename, config); err != nil {
				return err
			}
			tplReports[i] = &TplReport{
				report.Title,
				makeTagLinks(report.Tags),
				report.Stamp.Format(time.RFC822),
				reportFilename,
			}
		}
		i++
	}
	tplData := TplData{
		tplReports,
		makeHtml("reports"),
	}

	outputFilename := filepath.Join(config.WWWDir, "reports", "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateProgressPage(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
  <head>
		{{ index .Html "head" }}
		<meta http-equiv="refresh" content="4">
    <title>Progress</title>
  </head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Progress</h1>

				<ul class="reports-progress">
				{{range .Experiments}}
					<li>
						<table class="table table-bordered">
						  <tr>
								<th class="report-progress-th">Date</th>
								<td>{{ .Stamp }}</td>
							</tr>
							{{if .Title}}
								<tr><th>Title</th><td>{{ .Title }}</td></tr>
							{{end}}
							{{if .Tags}}
								<tr>
									<th>Tags</th>
									<td>
										{{range $tag, $catLink := .Tags}}
											<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
										{{end}}
									</td>
								</tr>
							{{end}}
							<tr><th>Experiment filename</th><td>{{ .Filename }}</td></tr>
							<tr><th>Message</th><td>{{ .Msg }}</td></tr>
							<tr>
								<th>Status</th>
								<td class="status-{{ .Status }}">{{ .Status }}</td>
							</tr>
						</table>
					</li>
				{{end}}
				</ul>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplExperiment struct {
		Title    string
		Tags     map[string]string
		Stamp    string
		Filename string
		Status   string
		Msg      string
	}

	type TplData struct {
		Experiments []*TplExperiment
		Html        map[string]template.HTML
	}

	experiments, err := progressMonitor.GetExperiments()
	if err != nil {
		return err
	}

	tplExperiments := make([]*TplExperiment, len(experiments))

	for i, experiment := range experiments {
		tplExperiments[i] = &TplExperiment{
			experiment.Title,
			makeTagLinks(experiment.Tags),
			experiment.Stamp.Format(time.RFC822),
			experiment.ExperimentFilename,
			experiment.Status.String(),
			experiment.Msg,
		}
	}
	tplData := TplData{tplExperiments, makeHtml("progress")}

	outputFilename := filepath.Join(config.WWWDir, "progress", "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func generateTagPage(
	config *config.Config,
	tagName string,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>Reports for tag: {{ .Tag }}</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports for tag: {{ .Tag }}</h1>

				<ul class="reports">
					{{range .Reports}}
						<li>
							<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
							Date: {{ .Stamp }}
							Tags:
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
							{{end}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplReport struct {
		Title    string
		Tags     map[string]string
		Stamp    string
		Filename string
	}

	type TplData struct {
		Tag     string
		Reports []*TplReport
		Html    map[string]template.HTML
	}

	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	i := 0
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			if inStrings(tagName, report.Tags) {
				reportFilename := makeReportFilename(report.Stamp, report.Title)
				tplReports[i] = &TplReport{
					report.Title,
					makeTagLinks(report.Tags),
					report.Stamp.Format(time.RFC822),
					fmt.Sprintf("/reports/%s", reportFilename),
				}
				i++
			}
		}
	}
	tplReports = tplReports[:i]
	tplData := TplData{tagName, tplReports, makeHtml("tag")}
	fullTagDir := filepath.Join(
		config.WWWDir,
		"reports",
		"tag",
		escapeString(tagName),
	)

	if err := os.MkdirAll(fullTagDir, 0740); err != nil {
		return err
	}
	outputFilename := filepath.Join(fullTagDir, "index.html")
	return writeTemplate(outputFilename, tpl, tplData)
}

func inStrings(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func generateTagPages(config *config.Config) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	tagsSeen := make(map[string]bool)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			for _, tag := range report.Tags {
				if _, ok := tagsSeen[tag]; !ok {
					if err := generateTagPage(config, tag); err != nil {
						return err
					}
					tagsSeen[tag] = true
				}
			}
		}
	}
	return nil
}

func makeTagLinks(tags []string) map[string]string {
	links := make(map[string]string, len(tags))
	for _, tag := range tags {
		links[tag] = makeTagLink(tag)
	}
	return links
}

func makeTagLink(tag string) string {
	return fmt.Sprintf(
		"/reports/tag/%s/",
		escapeString(tag),
	)
}

var nonAlphaNumRegexp = regexp.MustCompile("[^[:alnum:]]")

func escapeString(s string) string {
	crc32 := strconv.FormatUint(
		uint64(crc32.Checksum([]byte(s), crc32.MakeTable(crc32.IEEE))),
		36,
	)
	newS := nonAlphaNumRegexp.ReplaceAllString(s, "")
	return fmt.Sprintf("%s_%s", newS, crc32)
}

func generateReport(
	_report *report.Report,
	reportFilename string,
	config *config.Config,
) error {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>{{.Title}}</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>{{.Title}}</h1>

				<h2>Config</h2>
				<table class="neat-table">
					<tr class="title">
						<th> </th>
						<th class="last-column"> </th>
					</tr>
					<tr>
						<td>Tags</td>
						<td class="last-column">
							{{range $tag, $catLink := .Tags}}
								<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
							{{end}}<br />
						</td>
					</tr>
					<tr>
						<td>Number of records</td>
						<td class="last-column">{{.NumRecords}}</td>
					</tr>
					<tr>
						<td>Experiment file</td>
						<td class="last-column">{{.ExperimentFilename}}</td>
					</tr>
				</table>

				<table class="neat-table">
					<tr class="title">
						<th>Sort Order</th><th class="last-column">Direction</th>
					</tr>
					{{range .SortOrder}}
						<tr>
							<td>{{ .Field }}</td><td class="last-column">{{ .Direction }}</td>
						</tr>
					{{end}}
				</table>
			</div>

			<div class="container">
				<h2>Results</h2>
			</div>
			{{range .Assessments}}
				<div class="container">
					<h3>{{ .Rule }}</h3>

					<div class="pull-left aggregators">
						<table class="neat-table">
							<tr class="title">
								<th>Aggregator</th>
								<th>Value</th>
								<th class="last-column">Improvement</th>
							</tr>
							{{ range .Aggregators }}
							<tr>
								<td>{{ .Name }}</td>
								<td>{{ .Value }}</td>
								<td class="last-column">{{ .Difference }}</td>
							</tr>
							{{ end }}
						</table>
					</div>

					<div class="pull-left">
						<table class="neat-table">
							<tr class="title">
								<th>Goal</th><th class="last-column">Value</th>
							</tr>
							{{ range .Goals }}
							<tr>
								<td>{{ .Expr }}</td><td class="last-column">{{ .Passed }}</td>
							</tr>
							{{ end }}
						</table>
					</div>

				</div>
			{{ end }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`

	type TplData struct {
		Title              string
		Tags               map[string]string
		Stamp              string
		ExperimentFilename string
		NumRecords         int64
		SortOrder          []experiment.SortField
		Assessments        []*report.Assessment
		Html               map[string]template.HTML
	}

	tagLinks := makeTagLinks(_report.Tags)

	tplData := TplData{
		_report.Title,
		tagLinks,
		_report.Stamp.Format(time.RFC822),
		_report.ExperimentFilename,
		_report.NumRecords,
		_report.SortOrder,
		_report.Assessments,
		makeHtml("reports"),
	}

	fullReportFilename := filepath.Join(config.WWWDir, "reports", reportFilename)
	if err := writeTemplate(fullReportFilename, tpl, tplData); err != nil {
		return err
	}
	return nil
}

func writeTemplate(
	filename string,
	tpl string,
	tplData interface{},
) error {
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, tplData); err != nil {
		return err
	}
	return nil
}

func makeReportFilename(stamp time.Time, title string) string {
	timeSeconds := strconv.FormatInt(stamp.Unix(), 36)
	escapedTitle := escapeString(title)
	return fmt.Sprintf("%s_%s.html", escapedTitle, timeSeconds)
}

func countFiles(files []os.FileInfo) int {
	numFiles := 0
	for _, file := range files {
		if !file.IsDir() {
			numFiles++
		}
	}
	return numFiles
}

const htmlHead = `
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1">
<!-- The above 3 meta tags *must* come first in the head; any other head content must come *after* these tags -->

<link href="/css/bootstrap.min.css" rel="stylesheet">
<link href="/css/sitestyle.css" rel="stylesheet">

<!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
<!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
<!--[if lt IE 9]>
	<script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
	<script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
<![endif]-->`

const htmlBootstrapJS = `
		<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
			<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<!-- Include all compiled plugins (below), or include individual files as needed -->
			<script src="/js/bootstrap.min.js"></script>`

func makeHtmlNav(menuItem string) template.HTML {
	const tpl = `
<nav class="navbar navbar-inverse navbar-fixed-top">
	<div class="container">
		<div class="navbar-header">
			<button type="button" class="navbar-toggle collapsed"
			        data-toggle="collapse" data-target="#navbar"
			        aria-expanded="false" aria-controls="navbar">
				<span class="sr-only">Toggle navigation</span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
			</button>
			<a class="navbar-brand" href="/">RuleHunter</a>
		</div>

		<div id="navbar" class="collapse navbar-collapse">
			<ul class="nav navbar-nav">
				{{if eq .MenuItem "home"}}
					<li class="active"><a href="/">Home</a></li>
				{{else}}
					<li><a href="/">Home</a></li>
				{{end}}

				{{if eq .MenuItem "reports"}}
					<li class="active"><a href="/reports/">Reports</a></li>
				{{else}}
					<li><a href="/reports/">Reports</a></li>
				{{end}}

				{{if eq .MenuItem "tag"}}
					<li class="active"><a href=".">Tag</a></li>
				{{end}}

				{{if eq .MenuItem "progress"}}
					<li class="active"><a href="/progress/">Progress</a></li>
				{{else}}
					<li><a href="/progress/">Progress</a></li>
				{{end}}
			</ul>
		</div><!--/.nav-collapse -->
	</div>
</nav>`

	var doc bytes.Buffer
	validMenuItems := []string{
		"home",
		"reports",
		"tag",
		"progress",
	}

	foundValidItem := false
	for _, validMenuItem := range validMenuItems {
		if validMenuItem == menuItem {
			foundValidItem = true
		}
	}
	if !foundValidItem {
		panic(fmt.Sprintf("menuItem not valid: %s", menuItem))
	}

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}

	tplData := struct{ MenuItem string }{menuItem}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s",
			menuItem, err))
	}
	return template.HTML(doc.String())
}

func makeHtml(menuItem string) map[string]template.HTML {
	r := make(map[string]template.HTML)
	r["head"] = template.HTML(htmlHead)
	r["nav"] = makeHtmlNav(menuItem)
	r["bootstrapJS"] = template.HTML(htmlBootstrapJS)
	return r
}