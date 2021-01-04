/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package swampdisasm

import (
	"fmt"

	swampopcodeinst "github.com/swamp/opcodes/instruction"
	swampopcode "github.com/swamp/opcodes/opcode"
	swampopcodetype "github.com/swamp/opcodes/type"
)

type Register struct {
	id uint8
}

type Argument interface {
	String() string
}

type OpcodeInStream struct {
	position int
	octets   []byte
}

func NewOpcodeInStream(octets []byte) *OpcodeInStream {
	return &OpcodeInStream{octets: octets}
}

func (s *OpcodeInStream) IsEOF() bool {
	return s.position >= len(s.octets)
}

func (s *OpcodeInStream) readUint8() uint8 {
	if s.position == len(s.octets) {
		panic("swamp disassembler: read too far")
	}

	a := s.octets[s.position]

	s.position++

	return a
}

func (s *OpcodeInStream) readCommand() swampopcodeinst.Commands {
	return swampopcodeinst.Commands(s.readUint8())
}

func (s *OpcodeInStream) programCounter() swampopcodetype.ProgramCounter {
	return swampopcodetype.NewProgramCounter(uint16(s.position))
}

func (s *OpcodeInStream) readRegister() swampopcodetype.Register {
	return swampopcodetype.NewRegister(s.readUint8())
}

func (s *OpcodeInStream) readField() swampopcodetype.Field {
	return swampopcodetype.NewField(s.readUint8())
}

func (s *OpcodeInStream) readCount() int {
	return int(s.readUint8())
}

func (s *OpcodeInStream) readLabel() *swampopcodetype.Label {
	delta := uint16(s.readUint8())
	resultingPosition := s.programCounter().Add(delta)

	return swampopcodetype.NewLabelDefined("", resultingPosition)
}

func (s *OpcodeInStream) readLabelOffset(offset swampopcodetype.ProgramCounter) *swampopcodetype.Label {
	delta := uint16(s.readUint8())
	resultingPosition := offset.Add(delta)

	return swampopcodetype.NewLabelDefined("offset", resultingPosition)
}

func (s *OpcodeInStream) readRegisters() []swampopcodetype.Register {
	count := s.readCount()
	array := make([]swampopcodetype.Register, count)

	for i := 0; i < count; i++ {
		array[i] = s.readRegister()
	}

	return array
}

func disassembleListConj(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.ListConj {
	destination := s.readRegister()
	list := s.readRegister()
	item := s.readRegister()

	return swampopcodeinst.NewListConj(destination, item, list)
}

func disassembleListAppend(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.ListAppend {
	destination := s.readRegister()
	a := s.readRegister()
	b := s.readRegister()

	return swampopcodeinst.NewListAppend(destination, a, b)
}

func disassembleStringAppend(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.StringAppend {
	destination := s.readRegister()
	a := s.readRegister()
	b := s.readRegister()

	return swampopcodeinst.NewStringAppend(destination, a, b)
}

func disassembleBinaryOperator(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.IntBinaryOperator {
	destination := s.readRegister()
	a := s.readRegister()
	b := s.readRegister()

	return swampopcodeinst.NewIntBinaryOperator(cmd, destination, a, b)
}

func disassembleBitwiseOperator(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.IntBinaryOperator {
	destination := s.readRegister()
	a := s.readRegister()
	b := s.readRegister()

	return swampopcodeinst.NewIntBinaryOperator(cmd, destination, a, b)
}

func disassembleBitwiseUnaryOperator(cmd swampopcodeinst.Commands, s *OpcodeInStream) *swampopcodeinst.IntUnaryOperator {
	destination := s.readRegister()
	a := s.readRegister()

	return swampopcodeinst.NewIntUnaryOperator(cmd, destination, a)
}

func disassembleCreateStruct(s *OpcodeInStream) *swampopcodeinst.CreateStruct {
	destination := s.readRegister()
	arguments := s.readRegisters()

	return swampopcodeinst.NewCreateStruct(destination, arguments)
}

func disassembleCreateList(s *OpcodeInStream) *swampopcodeinst.CreateList {
	destination := s.readRegister()
	arguments := s.readRegisters()

	return swampopcodeinst.NewCreateList(destination, arguments)
}

func disassembleCall(s *OpcodeInStream) *swampopcodeinst.Call {
	destination := s.readRegister()
	functionRegister := s.readRegister()
	arguments := s.readRegisters()

	return swampopcodeinst.NewCall(destination, functionRegister, arguments)
}

func disassembleCallExternal(s *OpcodeInStream) *swampopcodeinst.CallExternal {
	destination := s.readRegister()
	functionRegister := s.readRegister()
	arguments := s.readRegisters()

	return swampopcodeinst.NewCallExternal(destination, functionRegister, arguments)
}

func disassembleCurry(s *OpcodeInStream) *swampopcodeinst.Curry {
	destination := s.readRegister()
	functionRegister := s.readRegister()
	arguments := s.readRegisters()

	return swampopcodeinst.NewCurry(destination, functionRegister, arguments)
}

func disassembleCreateEnum(s *OpcodeInStream) *swampopcodeinst.Enum {
	destination := s.readRegister()
	enumFieldIndex := s.readCount()
	arguments := s.readRegisters()

	return swampopcodeinst.NewEnum(destination, enumFieldIndex, arguments)
}

func disassembleUpdateStruct(s *OpcodeInStream) *swampopcodeinst.UpdateStruct {
	destination := s.readRegister()
	source := s.readRegister()
	count := s.readCount()

	var assignments []swampopcodeinst.CopyToFieldInfo

	for i := 0; i < count; i++ {
		fieldIndex := s.readField()
		sourceRegister := s.readRegister()
		assignment := swampopcodeinst.CopyToFieldInfo{Target: fieldIndex, Source: sourceRegister}
		assignments = append(assignments, assignment)
	}

	return swampopcodeinst.NewUpdateStruct(destination, source, assignments)
}

func disassembleGetStruct(s *OpcodeInStream) *swampopcodeinst.GetStruct {
	destination := s.readRegister()
	source := s.readRegister()
	count := s.readCount()

	var lookups []swampopcodetype.Field

	for i := 0; i < count; i++ {
		fieldIndex := s.readField()
		lookups = append(lookups, fieldIndex)
	}

	return swampopcodeinst.NewGetStruct(destination, source, lookups)
}

func disassembleCase(s *OpcodeInStream) *swampopcodeinst.EnumCase {
	destination := s.readRegister()
	source := s.readRegister()
	count := s.readCount()

	var jumps []swampopcodeinst.EnumCaseJump

	var lastLabel *swampopcodetype.Label

	for i := 0; i < count; i++ {
		enumValue := s.readUint8()
		argCount := s.readCount()

		var args []swampopcodetype.Register

		for j := 0; j < argCount; j++ {
			args = append(args, s.readRegister())
		}

		var label *swampopcodetype.Label

		if lastLabel != nil {
			label = s.readLabelOffset(lastLabel.DefinedProgramCounter())
		} else {
			label = s.readLabel()
		}

		lastLabel = label
		jump := swampopcodeinst.NewEnumCaseJump(enumValue, args, label)
		jumps = append(jumps, jump)
	}

	return swampopcodeinst.NewEnumCase(destination, source, jumps)
}

func disassembleRegCopy(s *OpcodeInStream) *swampopcodeinst.RegCopy {
	destination := s.readRegister()
	source := s.readRegister()

	return swampopcodeinst.NewRegCopy(destination, source)
}

func disassembleTailCall(s *OpcodeInStream) *swampopcodeinst.TailCall {
	return nil
}

func disassembleReturn(s *OpcodeInStream) *swampopcodeinst.Return {
	return swampopcodeinst.NewReturn()
}

func disassembleJump(s *OpcodeInStream) *swampopcodeinst.Jump {
	label := s.readLabel()

	return swampopcodeinst.NewJump(label)
}

func disassembleBranchFalse(s *OpcodeInStream) *swampopcodeinst.BranchFalse {
	test := s.readRegister()
	label := s.readLabel()

	return swampopcodeinst.NewBranchFalse(test, label)
}

func disassembleBranchTrue(s *OpcodeInStream) *swampopcodeinst.BranchTrue {
	test := s.readRegister()
	label := s.readLabel()

	return swampopcodeinst.NewBranchTrue(test, label)
}

func decodeOpcode(cmd swampopcodeinst.Commands, s *OpcodeInStream) swampopcode.Instruction {
	switch cmd {
	case swampopcodeinst.CmdAdd:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdSub:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdDiv:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdMul:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdEqual:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdNotEqual:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdLess:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdLessOrEqual:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdGreater:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdGreaterOrEqual:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdFixedDiv:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdFixedMul:
		return disassembleBinaryOperator(cmd, s)
	case swampopcodeinst.CmdListConj:
		return disassembleListConj(cmd, s)
	case swampopcodeinst.CmdListAppend:
		return disassembleListAppend(cmd, s)
	case swampopcodeinst.CmdStringAppend:
		return disassembleStringAppend(cmd, s)
	case swampopcodeinst.CmdCreateStruct:
		return disassembleCreateStruct(s)
	case swampopcodeinst.CmdCreateList:
		return disassembleCreateList(s)
	case swampopcodeinst.CmdUpdateStruct:
		return disassembleUpdateStruct(s)
	case swampopcodeinst.CmdStructGet:
		return disassembleGetStruct(s)
	case swampopcodeinst.CmdEnumCase:
		return disassembleCase(s)
	case swampopcodeinst.CmdRegCopy:
		return disassembleRegCopy(s)
	case swampopcodeinst.CmdCall:
		return disassembleCall(s)
	case swampopcodeinst.CmdCallExternal:
		return disassembleCallExternal(s)
	case swampopcodeinst.CmdTailCall:
		return disassembleTailCall(s)
	case swampopcodeinst.CmdCurry:
		return disassembleCurry(s)
	case swampopcodeinst.CmdCreateEnum:
		return disassembleCreateEnum(s)
	case swampopcodeinst.CmdReturn:
		return disassembleReturn(s)
	case swampopcodeinst.CmdJump:
		return disassembleJump(s)
	case swampopcodeinst.CmdBranchFalse:
		return disassembleBranchFalse(s)
	case swampopcodeinst.CmdBranchTrue:
		return disassembleBranchTrue(s)
	case swampopcodeinst.CmdBitwiseAnd:
		return disassembleBitwiseOperator(cmd, s)
	case swampopcodeinst.CmdBitwiseOr:
		return disassembleBitwiseOperator(cmd, s)
	case swampopcodeinst.CmdBitwiseXor:
		return disassembleBitwiseOperator(cmd, s)
	case swampopcodeinst.CmdBitwiseNot:
		return disassembleBitwiseUnaryOperator(cmd, s)
	case swampopcodeinst.CmdLogicalNot:
		return disassembleBitwiseUnaryOperator(cmd, s)
	}

	panic(fmt.Sprintf("swamp disassembler: unknown opcode:%v", cmd))

	//return nil
}

func Disassemble(octets []byte) []string {
	var lines []string

	s := NewOpcodeInStream(octets)

	for !s.IsEOF() {
		startPc := s.programCounter()
		cmd := s.readCommand()

		// fmt.Printf("disasembling :%s (%02x)\n", swampopcode.OpcodeToName(cmd), cmd)
		args := decodeOpcode(cmd, s)
		line := fmt.Sprintf("%02x: %v", startPc.Value(), args)
		lines = append(lines, line)
	}

	return lines
}
