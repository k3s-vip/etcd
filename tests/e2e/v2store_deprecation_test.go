// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/tests/v3/framework/e2e"
)

func createV2store(t testing.TB, dataDirPath string) {
	t.Log("Creating not-yet v2-deprecated etcd")

	cfg := e2e.ConfigStandalone(e2e.EtcdProcessClusterConfig{EnableV2: true, DataDirPath: dataDirPath, SnapshotCount: 5})
	epc, err := e2e.NewEtcdProcessCluster(t, cfg)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, epc.Stop())
	}()

	// We need to exceed 'SnapshotCount' such that v2 snapshot is dumped.
	for i := 0; i < 10; i++ {
		if err := e2e.CURLPut(epc, e2e.CURLReq{
			Endpoint: "/v2/keys/foo", Value: "bar" + fmt.Sprint(i),
			Expected: `{"action":"set","node":{"key":"/foo","value":"bar` + fmt.Sprint(i)}); err != nil {
			t.Fatalf("failed put with curl (%v)", err)
		}
	}
}

func assertVerifyCanStartV2deprecationNotYet(t testing.TB, dataDirPath string) {
	t.Log("verify: possible to start etcd with --v2-deprecation=not-yet mode")

	cfg := e2e.ConfigStandalone(e2e.EtcdProcessClusterConfig{EnableV2: true, DataDirPath: dataDirPath, V2deprecation: "not-yet", KeepDataDir: true})
	epc, err := e2e.NewEtcdProcessCluster(t, cfg)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, epc.Stop())
	}()

	if err := e2e.CURLGet(epc, e2e.CURLReq{
		Endpoint: "/v2/keys/foo",
		Expected: `{"action":"get","node":{"key":"/foo","value":"bar9","modifiedIndex":13,"createdIndex":13}}`}); err != nil {
		t.Fatalf("failed get with curl (%v)", err)
	}

}

func assertVerifyCannotStartV2deprecationWriteOnly(t testing.TB, dataDirPath string) {
	t.Log("Verify its infeasible to start etcd with --v2-deprecation=write-only mode")
	proc, err := e2e.SpawnCmd([]string{e2e.BinDir + "/etcd", "--v2-deprecation=write-only", "--data-dir=" + dataDirPath}, nil)
	assert.NoError(t, err)

	_, err = proc.Expect("detected disallowed custom content in v2store for stage --v2-deprecation=write-only")
	assert.NoError(t, err)
}

func TestV2Deprecation(t *testing.T) {
	e2e.BeforeTest(t)
	dataDirPath := t.TempDir()

	t.Run("create-storev2-data", func(t *testing.T) {
		createV2store(t, dataDirPath)
	})

	t.Run("--v2-deprecation=write-only fails", func(t *testing.T) {
		assertVerifyCannotStartV2deprecationWriteOnly(t, dataDirPath)
	})

	t.Run("--v2-deprecation=not-yet succeeds", func(t *testing.T) {
		assertVerifyCanStartV2deprecationNotYet(t, dataDirPath)
	})

}

func TestV2DeprecationWriteOnlyNoV2Api(t *testing.T) {
	e2e.BeforeTest(t)
	proc, err := e2e.SpawnCmd([]string{e2e.BinDir + "/etcd", "--v2-deprecation=write-only", "--enable-v2"}, nil)
	assert.NoError(t, err)

	_, err = proc.Expect("--enable-v2 and --v2-deprecation=write-only are mutually exclusive")
	assert.NoError(t, err)
}

func TestV2DeprecationCheckCustomContentOffline(t *testing.T) {
	e2e.BeforeTest(t)

	t.Run("WithCustomContent", func(t *testing.T) {
		dataDirPath := t.TempDir()

		createV2store(t, dataDirPath)

		assertVerifyCheckCustomContentOffline(t, dataDirPath)
	})

	t.Run("WithoutCustomContent", func(t *testing.T) {
		dataDirPath := ""

		func() {
			cCtx := getDefaultCtlCtx(t)

			cfg := cCtx.cfg
			cfg.ClusterSize = 3
			cfg.SnapshotCount = 5
			cfg.EnableV2 = true

			// create a cluster with 3 members
			epc, err := e2e.NewEtcdProcessCluster(t, &cfg)
			assert.NoError(t, err)

			cCtx.epc = epc
			dataDirPath = epc.Procs[0].Config().DataDirPath

			defer func() {
				assert.NoError(t, epc.Stop())
			}()

			// create key-values with v3 api
			for i := 0; i < 10; i++ {
				assert.NoError(t, ctlV3Put(cCtx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), ""))
			}
		}()

		proc, err := e2e.SpawnCmd([]string{e2e.BinDir + "/etcdutl", "check", "v2store", "--data-dir=" + dataDirPath}, nil)
		assert.NoError(t, err)

		_, err = proc.Expect("No custom content found in v2store")
		assert.NoError(t, err)
	})
}

func assertVerifyCheckCustomContentOffline(t *testing.T, dataDirPath string) {
	t.Logf("Checking custom content in v2store - %s", dataDirPath)

	proc, err := e2e.SpawnCmd([]string{e2e.BinDir + "/etcdutl", "check", "v2store", "--data-dir=" + dataDirPath}, nil)
	assert.NoError(t, err)

	_, err = proc.Expect("detected custom content in v2store")
	assert.NoError(t, err)
}
