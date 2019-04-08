package main

import (
  "fmt"
  "github.com/FLAGlab/DCoPN/communication"
)
//type Arc petrinet.Arc
func main() {
	//pId := 123
	//fpt := []Arc{Arc{p: 1, t: 1, w: 1}, Arc{p: 2, t: 2, w: 1}, Arc{p: 3, t: 2, w: 1}}
	//ftp := []Arc{Arc{t: 1, p: 2, w: 1}, Arc{t: 1, p: 3, w: 1}, Arc{t: 2, p: 4, w: 1}}
	//m := make(map[int]int)
	//inhi := []Arc{Arc{t: 2, p: 4, w:1}}
	//m[1] = 1
	//m[3] = 2
	//m[4] = 1
	//p := petrinet.InitPetriNet(4, 2, fpt, ftp, inhi, m)
	//fmt.Printf("%v", p)
	//p.Run()
	// petrinet.Test()
  fmt.Println("Starting....")
  communication.Run()
}
/*
1) Estados en conexion y desconexion (reachability graph -> hacerlo bien es tema de otra tesis)
2) Resolver conflictos

Tabla
| situacion   | Solucion  | attin   |
*/

/*

TODO: posibilidad de no disparar ninguna

*/
