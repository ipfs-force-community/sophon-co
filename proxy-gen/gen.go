package gen

import (
	"bytes"
	"fmt"
	"go/build"
	"reflect"
	"strings"

	"golang.org/x/tools/imports"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

// Gen generates the impl code for given api interface
func Gen(pkgName, structName string, api interface{}) ([]byte, error) {
	gen := newGenerator(pkgName, structName)
	if err := gen.register(reflect.TypeOf(api)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gen.write(&buf)

	// return format.Source(buf.Bytes())
	return imports.Process("", buf.Bytes(), nil)
}

type api struct {
	typ     reflect.Type
	methods []*method
}

func (a *api) writeDef(structName string, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("// impl %s\n", a.typ))
	for _, meth := range a.methods {
		meth.writeMethodDef(structName, buf)
	}
	buf.WriteString("\n\n")
}

type method struct {
	name      string
	in        []*genType
	out       []*genType
	returnErr bool
}

func (m method) writeMethodDef(structName string, buf *bytes.Buffer) {
	inDefs := make([]string, 0, len(m.in))
	inNames := make([]string, 0, len(m.in))
	outDefs := make([]string, 0, len(m.out))

	for i := range m.in {
		inName := fmt.Sprintf("in%d", i)
		def := fmt.Sprintf("%s %s", inName, m.in[i])

		inNames = append(inNames, inName)
		inDefs = append(inDefs, def)
	}

	for i := range m.out {
		def := fmt.Sprintf("out%d %s", i, m.out[i])
		outDefs = append(outDefs, def)
	}

	if m.returnErr {
		outDefs = append(outDefs, "err error")
	}

	buf.WriteString(fmt.Sprintf("func (p *%s) %s(%s) (%s) {\n", structName, m.name, strings.Join(inDefs, ", "), strings.Join(outDefs, ", ")))
	buf.WriteString(fmt.Sprintf(`cli, err := p.Select()
	if err != nil {
		err = fmt.Errorf("api %s %%v", err)
		return
	}
	`, m.name))
	buf.WriteString(fmt.Sprintf("return cli.%s(%s)", m.name, strings.Join(inNames, ", ")))
	buf.WriteString("}\n\n")
}

func newGenerator(pname string, sname string) *generator {
	return &generator{
		pkgName:    pname,
		structName: sname,
		apis:       make([]api, 0),

		depCounter: map[string]int{},
		deps:       map[string]*depDef{},
		types:      map[reflect.Type]*genType{},
	}
}

type generator struct {
	pkgName    string
	structName string
	apis       []api

	depCounter map[string]int
	deps       map[string]*depDef
	types      map[reflect.Type]*genType
}

func (g *generator) write(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.pkgName))
	g.writeImports(buf)
	g.writeTypeAssertion(buf)
	g.writeInterfaceDef(buf)
	g.writeStructDef(buf)
	g.writeImpls(buf)
}

func (g *generator) writeImports(buf *bytes.Buffer) {
	buf.WriteString("import (\n")
	for path, def := range g.deps {
		if def.name != "" {
			buf.WriteString(def.name + " ")
		}

		buf.WriteString("\"")
		buf.WriteString(path)
		buf.WriteString("\"\n")
	}
	buf.WriteString(")\n\n")
}

func (g *generator) writeInterfaceDef(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("type %sAPI interface {\n", g.structName))
	for _, api := range g.apis {
		buf.WriteString(api.typ.String() + "\n")
	}
	buf.WriteString("}\n\n")
}

func (g *generator) writeTypeAssertion(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("var _ %sAPI = (*%s)(nil)\n", g.structName, g.structName))
}

func (g *generator) writeStructDef(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("type %s struct {\n", g.structName))
	buf.WriteString(fmt.Sprintf("Select func() (%sAPI, error)\n", g.structName))
	buf.WriteString("}\n\n")
}

func (g *generator) writeImpls(buf *bytes.Buffer) {
	for _, api := range g.apis {
		api.writeDef(g.structName, buf)
	}
}

func (g *generator) register(raw reflect.Type) error {
	if raw.Kind() != reflect.Ptr || raw.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("api should be a ptr to an interface, got %s", raw)
	}

	apiTyp := raw.Elem()
	if _, err := g.registerType(apiTyp); err != nil {
		return fmt.Errorf("register interface %s: %w", apiTyp, err)
	}

	numMeth := apiTyp.NumMethod()
	a := api{
		typ:     raw.Elem(),
		methods: make([]*method, 0, numMeth),
	}

	for i := 0; i < numMeth; i++ {
		meth := apiTyp.Method(i)
		m, err := g.genMethod(meth)
		if err != nil {
			return fmt.Errorf("gen #%d meth %s: %w", i, meth.Name, err)
		}

		a.methods = append(a.methods, m)
	}

	g.apis = append(g.apis, a)
	return nil
}

func (g *generator) genMethod(meth reflect.Method) (*method, error) {
	mtyp := meth.Type
	numIn := mtyp.NumIn()
	numOut := mtyp.NumOut()

	m := method{
		name:      meth.Name,
		in:        make([]*genType, 0, numIn),
		out:       make([]*genType, 0, numOut),
		returnErr: false,
	}

	for i := 0; i < numIn; i++ {
		inTyp := mtyp.In(i)
		gt, err := g.registerType(inTyp)
		if err != nil {
			return nil, fmt.Errorf("register #%d in for %s: %w", i, mtyp, err)
		}

		m.in = append(m.in, gt)
	}

	for i := 0; i < numOut; i++ {
		outTyp := mtyp.Out(i)
		if i == numOut-1 && outTyp == errType {
			m.returnErr = true
			break
		}

		gt, err := g.registerType(outTyp)
		if err != nil {
			return nil, fmt.Errorf("register #%d out for %s: %w", i, mtyp, err)
		}

		m.out = append(m.out, gt)
	}

	return &m, nil
}

func (g *generator) registerType(t reflect.Type) (*genType, error) {
	if gt, ok := g.types[t]; ok {
		return gt, nil
	}

	gt, err := g.parseGenType(t)
	if err != nil {
		return nil, err
	}

	g.types[t] = gt
	return gt, nil
}

func (g *generator) parseGenType(raw reflect.Type) (*genType, error) {
	pkgPath := raw.PkgPath()
	if pkgPath != "" {
		dd, err := g.getDepDef(pkgPath)
		if err != nil {
			return nil, err
		}

		str := ""
		if dd.name != "" {
			str = strings.Replace(raw.String(), dd.origin+".", dd.name+".", 1)
		}
		return newGenType(raw, str), nil
	}

	kind := raw.Kind()
	switch kind {
	case reflect.Invalid, reflect.Func, reflect.UnsafePointer:
		return nil, fmt.Errorf("unexpected kind %s", kind)

	case reflect.Array:
		eleType, err := g.parseGenType(raw.Elem())
		if err != nil {
			return nil, fmt.Errorf("parse array element type for %s: %w", raw, err)
		}

		if eleType.basic() {
			return newGenType(raw, ""), nil
		}

		size := raw.Len()
		return newGenType(raw, fmt.Sprintf("[%d]%s", size, eleType)), nil

	case reflect.Chan:
		eleType, err := g.parseGenType(raw.Elem())
		if err != nil {
			return nil, fmt.Errorf("parse chan element type for %s: %w", raw, err)
		}

		if eleType.basic() {
			return newGenType(raw, ""), nil
		}

		chanWithDir := "chan"
		switch raw.ChanDir() {
		case reflect.SendDir:
			chanWithDir = "chan<-"

		case reflect.RecvDir:
			chanWithDir = "<-chan"

		case reflect.BothDir:

		}

		return newGenType(raw, fmt.Sprintf("%s %s", chanWithDir, eleType)), nil

	case reflect.Map:
		keyType, err := g.parseGenType(raw.Key())
		if err != nil {
			return nil, fmt.Errorf("parse map key type for %s: %w", raw, err)
		}

		valType, err := g.parseGenType(raw.Elem())
		if err != nil {
			return nil, fmt.Errorf("parse map value type for %s: %w", raw, err)
		}

		if keyType.basic() && valType.basic() {
			return newGenType(raw, ""), nil
		}

		return newGenType(raw, fmt.Sprintf("map[%s]%s", keyType, valType)), nil

	case reflect.Ptr:
		eleType, err := g.parseGenType(raw.Elem())
		if err != nil {
			return nil, fmt.Errorf("parse ptr type for %s: %w", raw, err)
		}

		if eleType.basic() {
			return newGenType(raw, ""), nil
		}

		return newGenType(raw, fmt.Sprintf("*%s", eleType)), nil

	case reflect.Slice:
		eleType, err := g.parseGenType(raw.Elem())
		if err != nil {
			return nil, fmt.Errorf("parse slice element type for %s: %w", raw, err)
		}

		if eleType.basic() {
			return newGenType(raw, ""), nil
		}

		return newGenType(raw, fmt.Sprintf("[]%s", eleType)), nil

	default:
		return newGenType(raw, ""), nil
	}
}

func (g *generator) getDepDef(path string) (*depDef, error) {
	if dd, ok := g.deps[path]; ok {
		return dd, nil
	}

	pkgName, err := getPkgName(path)
	if err != nil {
		return nil, err
	}

	dd := &depDef{
		path:   path,
		origin: pkgName,
		name:   "",
	}

	idx := g.depCounter[pkgName]
	g.depCounter[pkgName] = idx + 1

	if idx > 0 {
		dd.name = fmt.Sprintf("%s%d", pkgName, idx)
	}

	g.deps[path] = dd
	return dd, nil
}

type depDef struct {
	path   string
	origin string
	name   string
}

func newGenType(typ reflect.Type, str string) *genType {
	return &genType{
		raw: typ,
		str: str,
	}
}

type genType struct {
	raw reflect.Type
	str string
}

func (gt genType) String() string {
	if gt.str != "" {
		return gt.str
	}

	return gt.raw.String()
}

func (gt genType) basic() bool {
	return gt.str == ""
}

func getPkgName(path string) (string, error) {
	pkg, err := build.Import(path, ".", 0)
	if err != nil {
		return "", err
	}

	return pkg.Name, nil
}
