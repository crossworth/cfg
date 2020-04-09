package cfg

import (
	"reflect"
	"testing"
)

const completeExample = `name: "TestServer",
host: "0.0.0.0",
port: 7788,
players: 1024,
#password: "verysecurepassword", # remove hashtag before password to enable
announce: false, # set to false during development
#token: no-token, # only needed when announce: true
gamemode: "Freeroam",
website: "test.com",
language: "en",
description: "test",
debug: false, # set to true during development
useEarlyAuth: true,
earlyAuthUrl: 'https://login.example.com:PORT',
useCdn: true,
cdnUrl: 'https://cdn.example.com:PORT',
modules: [
  "node-module",
  "csharp-module"
],
resources: [
  "example"
],
tags: [ 
  "customTag1",
  "customTag2",
  "customTag3",
  "customTag4"
],
voice: {
  bitrate: 64000
  #externalSecret: 3499211612
  externalHost: localhost
  externalPort: 7798
  externalPublicHost: 94.19.213.159
  externalPublicPort: 7799
}`

func TestUnmarshal(t *testing.T) {
	t.Run("invalid pointer", func(t *testing.T) {
		err := Unmarshal([]byte(""), 10)
		if err == nil {
			t.Fail()
		}
	})

	t.Run("nil value", func(t *testing.T) {
		err := Unmarshal([]byte(""), nil)
		if err == nil {
			t.Fail()
		}
	})

	t.Run("simple value", func(t *testing.T) {
		v := struct {
			Value string `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : 'test',"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != "test" {
			t.Fatalf("wrong value decoded, expected 'test', got %q", v.Value)
		}
	})

	t.Run("simple value without quote", func(t *testing.T) {
		v := struct {
			Value string `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : test,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != "test" {
			t.Fatalf("wrong value decoded, expected 'test', got %q", v.Value)
		}
	})

	t.Run("simple value with comment", func(t *testing.T) {
		v := struct {
			Value string `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : test # comment,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != "test" {
			t.Fatalf("wrong value decoded, expected 'test', got %q", v.Value)
		}
	})

	t.Run("bool value", func(t *testing.T) {
		v := struct {
			Value bool `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : true,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if !v.Value {
			t.Fatalf("wrong value decoded, expected true, got false")
		}
	})

	t.Run("int value", func(t *testing.T) {
		v := struct {
			Value int `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : -100,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != -100 {
			t.Fatalf("wrong value decoded, expected -100, got %d", v.Value)
		}
	})

	t.Run("field without name", func(t *testing.T) {
		v := struct {
			Value int
		}{}

		err := Unmarshal([]byte("Value : -100,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != -100 {
			t.Fatalf("wrong value decoded, expected -100, got %d", v.Value)
		}
	})

	t.Run("field with float", func(t *testing.T) {
		v := struct {
			Value float32 `cfg:"value"`
		}{}

		err := Unmarshal([]byte("value : 10.5,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != 10.5 {
			t.Fatalf("wrong value decoded, expected 10.5, got %f", v.Value)
		}
	})

	t.Run("ignore field with -", func(t *testing.T) {
		v := struct {
			Value int `cfg:"-"`
		}{
			Value: 100,
		}

		err := Unmarshal([]byte("value : 10.5,"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Value != 100 {
			t.Fatalf("wrong value decoded, expected 100, got %d", v.Value)
		}
	})

	t.Run("multiple lines", func(t *testing.T) {
		v := struct {
			Name string `cfg:"name"`
			Host string `cfg:"host"`
		}{}

		err := Unmarshal([]byte(`name: "TestServer",
host: "0.0.0.0",`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Name != "TestServer" {
			t.Fatalf("wrong value decoded, expected TestServer, got %s", v.Name)
		}

		if v.Host != "0.0.0.0" {
			t.Fatalf("wrong value decoded, expected 0.0.0.0, got %s", v.Host)
		}
	})

	t.Run("multiple lines without comma", func(t *testing.T) {
		v := struct {
			Name string `cfg:"name"`
			Host string `cfg:"host"`
		}{}

		err := Unmarshal([]byte(`name: "TestServer"
host: "0.0.0.0"`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v.Name != "TestServer" {
			t.Fatalf("wrong value decoded, expected TestServer, got %s", v.Name)
		}

		if v.Host != "0.0.0.0" {
			t.Fatalf("wrong value decoded, expected 0.0.0.0, got %s", v.Host)
		}
	})

	t.Run("slice value", func(t *testing.T) {
		v := struct {
			Modules []string `cfg:"modules"`
		}{}

		err := Unmarshal([]byte(`modules: [
  "node-module",
  "csharp-module"
]`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if len(v.Modules) != 2 {
			t.Fatalf("wrong value decoded, expected Modules length to be 2, got %d", len(v.Modules))
		}

		if v.Modules[0] != "node-module" {
			t.Fatalf("wrong value decoded, expected first module to be node-module, got %s", v.Modules[0])
		}

		if v.Modules[1] != "csharp-module" {
			t.Fatalf("wrong value decoded, expected first module to be csharp-module, got %s", v.Modules[1])
		}
	})

	t.Run("slice value without comma", func(t *testing.T) {
		v := struct {
			Modules []string `cfg:"modules"`
		}{}

		err := Unmarshal([]byte(`modules: [
  "node-module"
  "csharp-module"
]`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if len(v.Modules) != 2 {
			t.Fatalf("wrong value decoded, expected Modules length to be 2, got %d", len(v.Modules))
		}

		if v.Modules[0] != "node-module" {
			t.Fatalf("wrong value decoded, expected first module to be node-module, got %s", v.Modules[0])
		}

		if v.Modules[1] != "csharp-module" {
			t.Fatalf("wrong value decoded, expected first module to be csharp-module, got %s", v.Modules[1])
		}
	})

	t.Run("slice value with new lines", func(t *testing.T) {
		v := struct {
			Modules []string `cfg:"modules"`
		}{}

		err := Unmarshal([]byte(`modules: 
[
  "node-module"

  "csharp-module"

]`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if len(v.Modules) != 2 {
			t.Fatalf("wrong value decoded, expected Modules length to be 2, got %d", len(v.Modules))
		}

		if v.Modules[0] != "node-module" {
			t.Fatalf("wrong value decoded, expected first module to be node-module, got %s", v.Modules[0])
		}

		if v.Modules[1] != "csharp-module" {
			t.Fatalf("wrong value decoded, expected first module to be csharp-module, got %s", v.Modules[1])
		}
	})

	t.Run("slice value in single line", func(t *testing.T) {
		v := struct {
			Modules []string `cfg:"modules"`
		}{}

		err := Unmarshal([]byte(`modules: [ "node-module", "csharp-module"]`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if len(v.Modules) != 2 {
			t.Fatalf("wrong value decoded, expected Modules length to be 2, got %d", len(v.Modules))
		}

		if v.Modules[0] != "node-module" {
			t.Fatalf("wrong value decoded, expected first module to be node-module, got %s", v.Modules[0])
		}

		if v.Modules[1] != "csharp-module" {
			t.Fatalf("wrong value decoded, expected first module to be csharp-module, got %s", v.Modules[1])
		}
	})

	t.Run("slice value in single line without space", func(t *testing.T) {
		v := struct {
			Modules []string `cfg:"modules"`
		}{}

		err := Unmarshal([]byte(`modules: ["node-module","csharp-module"]`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if len(v.Modules) != 2 {
			t.Fatalf("wrong value decoded, expected Modules length to be 2, got %d", len(v.Modules))
		}

		if v.Modules[0] != "node-module" {
			t.Fatalf("wrong value decoded, expected first module to be node-module, got %s", v.Modules[0])
		}

		if v.Modules[1] != "csharp-module" {
			t.Fatalf("wrong value decoded, expected first module to be csharp-module, got %s", v.Modules[1])
		}
	})

	t.Run("struct withing struct", func(t *testing.T) {
		v := struct {
			Value struct {
				Test bool `cfg:"test"`
			} `cfg:"value"`
		}{}

		err := Unmarshal([]byte(`value : {
	test: true
}`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if !v.Value.Test {
			t.Fatal("wrong value decoded, expected inner Test to be true, got false")
		}
	})

	t.Run("struct withing complex struct", func(t *testing.T) {
		v := struct {
			Value struct {
				Test  bool    `cfg:"test"`
				Str   string  `cfg:"string"`
				Float float64 `cfg:"float"`
			} `cfg:"value"`
		}{}

		err := Unmarshal([]byte(`value : {
	test: true,
	string: "this is a string",
	float: 3.14159265359
}`), &v)
		if err != nil {
			t.Fatal(err)
		}

		if !v.Value.Test {
			t.Fatal("wrong value decoded, expected inner Test to be true, got false")
		}

		if v.Value.Str != "this is a string" {
			t.Fatalf("wrong value decoded, expected inner Str to be \"this is a string\", got %q", v.Value.Str)
		}

		if v.Value.Float != 3.14159265359 {
			t.Fatalf("wrong value decoded, expected inner Float to be 3.14159265359, got %f", v.Value.Float)
		}
	})

	t.Run("complete example", func(t *testing.T) {
		v := struct {
			Name         string   `cfg:"name"`
			Host         string   `cfg:"host"`
			Port         int      `cfg:"port"`
			Players      int      `cfg:"players"`
			Announce     bool     `cfg:"announce"`
			GameMode     string   `cfg:"gamemode"`
			WebSite      string   `cfg:"website"`
			Language     string   `cfg:"language"`
			Description  string   `cfg:"description"`
			Debug        bool     `cfg:"debug"`
			UseEarlyAuth bool     `cfg:"useEarlyAuth"`
			EarlyAuthURL string   `cfg:"earlyAuthUrl"`
			UseCDN       bool     `cfg:"useCdn"`
			CDNUrl       string   `cfg:"cdnUrl"`
			Modules      []string `cfg:"modules"`
			Resources    []string `cfg:"resources"`
			Tags         []string `cfg:"tags"`
			Voice        struct {
				BitRate            int    `cfg:"bitrate"`
				ExternalHost       string `cfg:"externalHost"`
				ExternalPort       int    `cfg:"externalPort"`
				ExternalPublicHost string `cfg:"externalPublicHost"`
				ExternalPublicPort int    `cfg:"externalPublicPort"`
			} `cfg:"voice"`
		}{}

		err := Unmarshal([]byte(completeExample), &v)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func Test_extractFields(t *testing.T) {
	s := struct {
		F1 int
		F2 bool `cfg:"value_f2"`
		F3 []string
		F4 string
	}{}

	fields1 := extractFields(reflect.ValueOf(s))

	if len(fields1) != 4 {
		t.Fail()
	}

	testFields(t, fields1, []string{"F1", "value_f2", "F3", "F4"})

	fields2 := extractFields(reflect.ValueOf(&s))

	if len(fields2) != 4 {
		t.Fail()
	}

	testFields(t, fields2, []string{"F1", "value_f2", "F3", "F4"})
}

func testFields(t *testing.T, fields []field, tagsExpected []string) {
	t.Helper()

	for i := range fields {
		if fields[i].Tag != tagsExpected[i] {
			t.Fatalf("wrong tag, expected %q got %q", tagsExpected[i], fields[i].Tag)
		}
	}
}

func Test_setValue(t *testing.T) {
	s := struct {
		Value string
	}{}

	rv := reflect.ValueOf(&s)

	fields := extractFields(rv)

	if len(fields) != 1 {
		t.Fail()
	}

	err := setValue(fields[0], "someValue")
	if err != nil {
		t.Fatal(err)
	}

	if s.Value != "someValue" {
		t.Fatalf("wrong value, expected someValue, got %q", s.Value)
	}
}

func Test_setSliceValue(t *testing.T) {
	s := struct {
		Values []string
	}{}

	rv := reflect.ValueOf(&s)

	fields := extractFields(rv)

	if len(fields) != 1 {
		t.Fail()
	}

	err := setSliceValue(fields[0], "someValue")
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Values) != 1 {
		t.Fail()
	}

	err = setSliceValue(fields[0], "andAnotherOne")
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Values) != 2 {
		t.Fail()
	}
}

func TestMarshal(t *testing.T) {
	t.Run("encode simple values", func(t *testing.T) {
		v := struct {
			TestStr   string
			TestBool  bool
			TestInt   int
			TestUint  uint `cfg:"test_uint"`
			TestFloat float64
		}{
			TestStr:   "this is a string",
			TestBool:  true,
			TestInt:   -100,
			TestUint:  100,
			TestFloat: 3.14,
		}

		data, err := Marshal(v)
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != `TestStr: 'this is a string',
TestBool: true,
TestInt: -100,
test_uint: 100,
TestFloat: 3.140000` {
			t.Fatalf("wrong value enconded, got %q", string(data))
		}
	})

	t.Run("encode slice", func(t *testing.T) {
		v := struct {
			StringSlice []string
		}{
			StringSlice: []string{"a", "b", "c"},
		}

		data, err := Marshal(v)
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != `StringSlice: [
  'a',
  'b',
  'c'
]` {
			t.Fatalf("wrong value enconded, got %q", string(data))
		}
	})

	t.Run("encode inner struct", func(t *testing.T) {
		v := struct {
			StructSlice struct {
				Working bool
				IntVal  int
			} `cfg:"struct_slice"`
		}{
			StructSlice: struct {
				Working bool
				IntVal  int
			}{
				Working: true,
				IntVal:  42,
			},
		}

		data, err := Marshal(v)
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != `struct_slice: {
  Working: true,
  IntVal: 42
}` {
			t.Fatalf("wrong value enconded, got %q", string(data))
		}
	})
}
