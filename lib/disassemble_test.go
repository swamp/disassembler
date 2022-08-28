/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package swampdisasm_sp

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	s := "17000000000100000002000000000b00270000000002000000010006"

	octets, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	stringLines := Disassemble(octets, true)
	output := fmt.Sprintf("%v", stringLines)

	const expectedOutput = `[0000: not 0,1 0009: brfa 0 [label @001b] 0010: cpy 0,(2:1) 001b: ret]`

	fmt.Println(output)

	if output != expectedOutput {
		t.Errorf("disassemble produced wrong output. expected\n%s\nbut received\n%s\n", expectedOutput, output)
	}
}
