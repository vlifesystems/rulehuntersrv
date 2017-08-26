/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

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

const categoryTpl = `
<!DOCTYPE html>
<html>
	<head>
		{{ index .Html "head" }}
		<title>Reports for category: {{ .Category }}</title>
	</head>

	<body>
		{{ index .Html "nav" }}

		<div id="content">
			<div class="container">
				<h1>Reports for category: {{ .Category }}</h1>

				<ul class="reports">
					{{range .Reports}}
						<li>
							<a class="title" href="{{ .Filename }}">{{ .Title }}</a><br />
							Date: {{ .DateTime }} &nbsp;
							{{if .Category}}
								Category: <a href="{{ .CategoryURL }}">{{ .Category }}</a> &nbsp;
							{{end}}
							{{if .Tags}}
								Tags:
								{{range $tag, $catLink := .Tags}}
									<a href="{{ $catLink }}">{{ $tag }}</a> &nbsp;
								{{end}}
							{{end}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>

		<div id="footer" class="container">
			{{ index .Html "footer" }}
		</div>

		{{ index .Html "bootstrapJS" }}
	</body>
</html>`