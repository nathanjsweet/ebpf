package ebpf

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
)

func TestLoadCollectionSpec(t *testing.T) {
	spec, err := LoadCollectionSpec("testdata/loader.elf")
	if err != nil {
		t.Fatal("Can't parse ELF:", err)
	}

	hashMapSpec := &MapSpec{Hash, 4, 2, 42, 4242, nil}
	checkMapSpec(t, spec.Maps, "hash_map", hashMapSpec)
	checkMapSpec(t, spec.Maps, "array_of_hash_map", &MapSpec{
		ArrayOfMaps, 4, 0, 2, 0, hashMapSpec,
	})

	hashMap2Spec := &MapSpec{Hash, 2, 1, 21, 2121, nil}
	checkMapSpec(t, spec.Maps, "hash_map2", hashMap2Spec)
	checkMapSpec(t, spec.Maps, "hash_of_hash_map", &MapSpec{
		HashOfMaps, 4, 0, 2, 0, hashMap2Spec,
	})

	checkProgramSpec(t, spec.Programs, "xdp_prog", &ProgramSpec{
		Type:    XDP,
		License: "MIT",
	})
	checkProgramSpec(t, spec.Programs, "no_relocation", &ProgramSpec{
		Type:    SocketFilter,
		License: "MIT",
	})
}

func Test64bitImmediate(t *testing.T) {
	// r1 = math.MinInt32 - 1
	prog := []byte{
		0x18, 0x01, 0x00, 0x00, 0xff, 0xff, 0xff, 0x7f,
		0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff,
	}

	insns, _, err := loadInstructions(binary.LittleEndian, prog)
	if err != nil {
		t.Fatal(err)
	}

	if len(insns) != 1 {
		t.Fatal("Expected one instruction, got", len(insns))
	}

	if c := insns[0].Constant; c != math.MinInt32-1 {
		t.Errorf("Expected immediate to be %v, got %v", math.MinInt32-1, c)
	}
}

func checkMapSpec(t *testing.T, maps map[string]*MapSpec, name string, want *MapSpec) {
	t.Helper()

	have, ok := maps[name]
	if !ok {
		t.Errorf("Missing map %s", name)
		return
	}

	mapSpecEqual(t, name, have, want)
}

func mapSpecEqual(t *testing.T, name string, have, want *MapSpec) {
	t.Helper()

	if have.Type != want.Type {
		t.Errorf("%s: expected type %v, got %v", name, want.Type, have.Type)
	}

	if have.KeySize != want.KeySize {
		t.Errorf("%s: expected key size %v, got %v", name, want.KeySize, have.KeySize)
	}

	if have.ValueSize != want.ValueSize {
		t.Errorf("%s: expected value size %v, got %v", name, want.ValueSize, have.ValueSize)
	}

	if have.MaxEntries != want.MaxEntries {
		t.Errorf("%s: expected max entries %v, got %v", name, want.MaxEntries, have.MaxEntries)
	}

	if have.Flags != want.Flags {
		t.Errorf("%s: expected flags %v, got %v", name, want.Flags, have.Flags)
	}

	switch {
	case have.InnerMap != nil && want.InnerMap == nil:
		t.Errorf("%s: extraneous InnerMap", name)
	case have.InnerMap == nil && want.InnerMap != nil:
		t.Errorf("%s: missing InnerMap", name)
	case have.InnerMap != nil && want.InnerMap != nil:
		mapSpecEqual(t, name+".InnerMap", have.InnerMap, want.InnerMap)
	}
}

func checkProgramSpec(t *testing.T, progs map[string]*ProgramSpec, name string, want *ProgramSpec) {
	t.Helper()

	have, ok := progs[name]
	if !ok {
		t.Fatalf("Missing program %s", name)
		return
	}

	if have.License != want.License {
		t.Errorf("%s: expected %v license, got %v", name, want.License, have.License)
	}

	if have.Type != want.Type {
		t.Errorf("%s: expected %v program, got %v", name, want.Type, have.Type)
	}

	if want.Instructions != nil && !reflect.DeepEqual(have.Instructions, want.Instructions) {
		t.Log("Expected program")
		t.Log(want.Instructions)
		t.Log("Actual program")
		t.Log(want.Instructions)
		t.Error("Instructions do not match")
	}
}
