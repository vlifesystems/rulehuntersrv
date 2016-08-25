/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
)

// And represents a rule determening if ruleA AND ruleB
type And struct {
	ruleA Rule
	ruleB Rule
}

func NewAnd(ruleA Rule, ruleB Rule) Rule {
	return &And{ruleA: ruleA, ruleB: ruleB}
}

func (r *And) String() string {
	// TODO: Consider making this AND rather than &&
	return fmt.Sprintf("%s && %s", r.ruleA, r.ruleB)
}

func (r *And) GetInNiParts() (bool, string, string) {
	return false, "", ""
}

func (r *And) IsTrue(record ddataset.Record) (bool, error) {
	lh, err := r.ruleA.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	rh, err := r.ruleB.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	return lh && rh, nil
}