package dao

import (
	"github.com/xuhaojun/chipmunk"
)

func ToCpBodyClient(body *chipmunk.Body) *CpBodyClient {
	// TODO
	// handle more shape
	shapeClients := make([]interface{}, len(body.Shapes))
	for i, shape := range body.Shapes {
		var shapeClient map[string]interface{}
		switch realShape := shape.ShapeClass.(type) {
		case *chipmunk.CircleShape:
			shapeClient = map[string]interface{}{
				"type":   "circle",
				"group":  shape.Group,
				"layer":  shape.Layer,
				"radius": realShape.Radius,
				"sensor": shape.IsSensor,
				"position": &CpVectClient{
					realShape.Position.X,
					realShape.Position.Y,
				},
			}
		case *chipmunk.BoxShape:
			shapeClient = map[string]interface{}{
				"type":   "box",
				"group":  shape.Group,
				"layer":  shape.Layer,
				"width":  realShape.Width,
				"height": realShape.Height,
				"sensor": shape.IsSensor,
				"position": &CpVectClient{
					realShape.Position.X,
					realShape.Position.Y,
				},
			}
		case *chipmunk.SegmentShape:
			shapeClient = map[string]interface{}{
				"type":   "segment",
				"group":  shape.Group,
				"layer":  shape.Layer,
				"radius": realShape.Radius,
				"sensor": shape.IsSensor,
				"a": &CpVectClient{
					realShape.A.X,
					realShape.A.Y,
				},
				"b": &CpVectClient{
					realShape.B.X,
					realShape.B.Y,
				},
			}
		}
		shapeClients[i] = shapeClient
	}
	var cpBodyClient *CpBodyClient
	pos := body.Position()
	vel := body.Velocity()
	if body.IsStatic() {
		cpBodyClient = &CpBodyClient{
			Shapes: shapeClients,
		}
	} else {
		cpBodyClient = &CpBodyClient{
			Mass:  body.Mass(),
			Angle: body.Angle(),
			Position: &CpVectClient{
				pos.X,
				pos.Y,
			},
			Velocity: &CpVectClient{
				vel.X,
				vel.Y,
			},
			Shapes: shapeClients,
		}
	}
	return cpBodyClient
}
