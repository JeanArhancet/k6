/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2021 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package execution

import (
	"context"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/stats"
)

func TestVUTags(t *testing.T) {
	t.Parallel()

	rt := goja.New()
	ctx := common.WithRuntime(context.Background(), rt)
	ctx = lib.WithState(ctx, &lib.State{
		Options: lib.Options{
			SystemTags: stats.NewSystemTagSet(stats.TagVU),
		},
		Tags: lib.NewTagMap(map[string]string{
			"vu": "42",
		}),
	})
	m, ok := New().NewModuleInstance(
		&modulestest.InstanceCore{
			Runtime: rt,
			InitEnv: &common.InitEnvironment{},
			Ctx:     ctx,
		},
	).(*ModuleInstance)
	require.True(t, ok)
	require.NoError(t, rt.Set("exec", m.GetExports().Default))

	// overwrite a system tag is allowed
	_, err := rt.RunString(`exec.vu.tags["vu"] = 101`)
	require.NoError(t, err)
	val, err := rt.RunString(`exec.vu.tags["vu"]`)
	require.NoError(t, err)
	// different type from string are implicitly converted into string
	assert.Equal(t, "101", val.String())

	// not found
	tag, err := rt.RunString(`exec.vu.tags["not-existing-tag"]`)
	require.NoError(t, err)
	assert.Equal(t, "undefined", tag.String())

	// set
	_, err = rt.RunString(`exec.vu.tags["custom-tag"] = "mytag"`)
	require.NoError(t, err)

	// get
	tag, err = rt.RunString(`exec.vu.tags["custom-tag"]`)
	require.NoError(t, err)
	assert.Equal(t, "mytag", tag.String())

	// json encoding
	encoded, err := rt.RunString(`JSON.stringify(exec.vu.tags)`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"vu":"101","custom-tag":"mytag"}`, encoded.String())
}
