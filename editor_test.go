package ebpf

import (
	"fmt"
	"testing"
)

func TestRewriteUint64(t *testing.T) {
	insns := Instructions{
		BPFILdImm64(Reg0, 0).Ref("ret"),
		BPFIOp(Exit),
	}

	ed := Edit(&insns)
	ed.RewriteUint64("ret", 42)

	spec := &ProgramSpec{
		Type:         XDP,
		Instructions: insns,
		License:      "MIT",
	}

	prog, err := NewProgram(spec)
	if err != nil {
		t.Fatal(err)
	}
	defer prog.Close()

	ret, _, err := prog.Test(make([]byte, 14))
	if err != nil {
		t.Fatal(err)
	}

	if ret != 42 {
		t.Error("Expected return value to be 42, got", ret)
	}
}

func ExampleEditor_RewriteUint64() {
	// The assembly is equivalent to this C:
	//
	//    unsigned long my_ret;
	//    unsigned long func() {
	//        return (int)my_ret;
	//    }
	//
	insns := Instructions{
		BPFILdImm64(Reg0, 0).Ref("my_ret"),
		BPFIOp(Exit),
	}

	editor := Edit(&insns)
	editor.RewriteUint64("my_ret", 42)

	fmt.Println(insns)

	// Output: 	0: LdImmDW dst: r0 imm: 42
	// 	2: Exit
}

func TestEditorLink(t *testing.T) {
	insns := Instructions{
		BPFIDstSrcImm(Call, Reg0, Reg1, -1).Ref("my_func"),
		BPFIOp(Exit),
	}

	editor := Edit(&insns)
	err := editor.Link(Instructions{
		BPFILdImm64(Reg0, 1337).Sym("my_func"),
		BPFIOp(Exit),
	})
	if err != nil {
		t.Fatal(err)
	}

	prog, err := NewProgram(&ProgramSpec{
		Type:         XDP,
		Instructions: insns,
		License:      "MIT",
	})
	if err != nil {
		t.Fatal(err)
	}

	ret, _, err := prog.Test(make([]byte, 14))
	if err != nil {
		t.Fatal(err)
	}

	if ret != 1337 {
		t.Errorf("Expected return code 1337, got %d", ret)
	}
}