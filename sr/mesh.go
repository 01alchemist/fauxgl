package sr

import (
	"math"
)

type Mesh struct {
	Triangles []*Triangle
	box       *Box
}

func NewMesh(triangles []*Triangle) *Mesh {
	return &Mesh{triangles, nil}
}

func (m *Mesh) dirty() {
	m.box = nil
}

func (m *Mesh) Copy() *Mesh {
	triangles := make([]*Triangle, len(m.Triangles))
	for i, t := range m.Triangles {
		a := *t
		triangles[i] = &a
	}
	return NewMesh(triangles)
}

func (a *Mesh) Add(b *Mesh) {
	a.Triangles = append(a.Triangles, b.Triangles...)
	a.dirty()
}

func (m *Mesh) BoundingBox() Box {
	if m.box == nil {
		box := BoxForTriangles(m.Triangles)
		m.box = &box
	}
	return *m.box
}

func smoothNormalsThreshold(normal Vector, normals []Vector, threshold float64) Vector {
	result := Vector{}
	for _, x := range normals {
		if x.Dot(normal) >= threshold {
			result = result.Add(x)
		}
	}
	return result.Normalize()
}

func (m *Mesh) SmoothNormalsThreshold(radians float64) {
	threshold := math.Cos(radians)
	lookup := make(map[Vector][]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1] = append(lookup[t.V1], t.N1)
		lookup[t.V2] = append(lookup[t.V2], t.N2)
		lookup[t.V3] = append(lookup[t.V3], t.N3)
	}
	for _, t := range m.Triangles {
		t.N1 = smoothNormalsThreshold(t.N1, lookup[t.V1], threshold)
		t.N2 = smoothNormalsThreshold(t.N2, lookup[t.V2], threshold)
		t.N3 = smoothNormalsThreshold(t.N3, lookup[t.V3], threshold)
	}
}

func (m *Mesh) SmoothNormals() {
	lookup := make(map[Vector]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1] = lookup[t.V1].Add(t.N1)
		lookup[t.V2] = lookup[t.V2].Add(t.N2)
		lookup[t.V3] = lookup[t.V3].Add(t.N3)
	}
	for k, v := range lookup {
		lookup[k] = v.Normalize()
	}
	for _, t := range m.Triangles {
		t.N1 = lookup[t.V1]
		t.N2 = lookup[t.V2]
		t.N3 = lookup[t.V3]
	}
}

func (m *Mesh) UnitCube() {
	m.FitInside(Box{Vector{}, Vector{1, 1, 1}}, Vector{})
	m.MoveTo(Vector{}, Vector{0.5, 0.5, 0.5})
}

func (m *Mesh) MoveTo(position, anchor Vector) {
	matrix := Translate(position.Sub(m.BoundingBox().Anchor(anchor)))
	m.Transform(matrix)
}

func (m *Mesh) FitInside(box Box, anchor Vector) {
	scale := box.Size().Div(m.BoundingBox().Size()).MinComponent()
	extra := box.Size().Sub(m.BoundingBox().Size().MulScalar(scale))
	matrix := Identity()
	matrix = matrix.Translate(m.BoundingBox().Min.Negate())
	matrix = matrix.Scale(Vector{scale, scale, scale})
	matrix = matrix.Translate(box.Min.Add(extra.Mul(anchor)))
	m.Transform(matrix)
}

func (m *Mesh) Transform(matrix Matrix) {
	for _, t := range m.Triangles {
		t.V1 = matrix.MulPosition(t.V1)
		t.V2 = matrix.MulPosition(t.V2)
		t.V3 = matrix.MulPosition(t.V3)
		t.N1 = matrix.MulDirection(t.N1)
		t.N2 = matrix.MulDirection(t.N2)
		t.N3 = matrix.MulDirection(t.N3)
	}
	m.dirty()
}

func (m *Mesh) SaveSTL(path string) error {
	return SaveSTL(path, m)
}
