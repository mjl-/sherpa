package sherpa

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mjl-/sherpadoc"
)

var testI int64
var testU uint64

func TestIntstr(t *testing.T) {
	tcheck := func(err error, action string) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %s\n", action, err)
		}
	}

	i := new(Int64s)
	*i = 1 << 62
	buf, err := json.Marshal(i)
	tcheck(err, "marshal Int64s")

	var i1 Int64s
	err = json.Unmarshal(buf, &i1)
	tcheck(err, "unmarshal Intstr64")
	if i.Int() != i1.Int() {
		t.Fatalf("int64str value mismatch, parsed %v != original %v", i1.Int(), i.Int())
	}

	api := testAPI{}
	h, err := NewHandler("/", "0.0.1", api, &sherpadoc.Section{}, nil)
	tcheck(err, "NewHandler")
	req := httptest.NewRequest("POST", "/int64Test", strings.NewReader(`{"params": ["-4611686018427387904", "4611686018427387904"]}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if testI != -4611686018427387904 {
		t.Errorf("bad Int64Test, for i got %d, expected -4611686018427387904", testI)
	}
	if testU != 4611686018427387904 {
		t.Errorf("bad Int64Test, for u got %d, expected 4611686018427387904", testU)
	}

	testI = -1
	testU = 1
	// TODO: null should not be accepted
	req = httptest.NewRequest("POST", "/int64Test", strings.NewReader(`{"params": [-4611686018427387904, null]}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Errorf("call failed, got %v, expected 200\n", resp.Code)
	}
	if testI != -4611686018427387904 {
		t.Errorf("bad Int64Test, for i got %d, expected -4611686018427387904", testI)
	}
	if testU != 0 {
		t.Errorf("bad Int64Test, for u got %d, expected 0", testU)
	}
	var response struct {
		Result Nums
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	tcheck(err, "parsing returned json")
	const intstrVal = -1 << 62
	const intVal = 1 << 62
	if response.Result.Intstr != intstrVal {
		t.Fatalf("bad response, got %v, expected %v (%s)\n", response.Result.Intstr, int64(intstrVal), resp.Body.String())
	}
	if response.Result.Int != intVal {
		t.Fatalf("bad response, got %v, expected %v\n", response.Result.Int, int64(intVal))
	}
}

type testAPI struct {
}

type Nums struct {
	Intstr int64 `json:",string"`
	Int    int64
}

func (t testAPI) Int64Test(i Int64s, u Uint64s) Nums {
	testI = int64(i)
	testU = uint64(u)
	return Nums{
		Intstr: -1 << 62,
		Int:    1 << 62,
	}
}
