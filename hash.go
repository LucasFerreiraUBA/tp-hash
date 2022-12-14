package diccionario

import (
	"fmt"
	TDALista "lista"
)

// CONSTANTES ----------------------------------------------
const (
	VACIO EstadoCelda = iota
	OCUPADO
)
const (
	CONSTANTE_HOPSCOTCH = 4
)

// TYPES --------------------------------------------------
type EstadoCelda int

type Celda[V any, K comparable] struct {
	estado    EstadoCelda
	valor     V
	clave     K
	hashingID int
}

type hashCerrado[K comparable, V any] struct {
	tabla    []Celda[V, K]
	cantidad int
	iter     TDALista.IteradorLista[Celda[V, K]]
}

type iteradorDiccionarioExterno[K comparable, V any] struct {
	iterador *TDALista.IteradorLista[Celda[V, K]]
}

// FUNCIONES CREADORAS: ----------------------

func crearCelda[V any, K comparable](clave K, valor V) *Celda[V, K] {
	celda := new(Celda[V, K])
	celda.valor = valor
	celda.clave = clave
	celda.estado = OCUPADO
	return celda
}

func CrearHash[K comparable, V any]() Diccionario[K, V] {
	hash := new(hashCerrado[K, V])
	hash.tabla = make([]Celda[V, K], 10)
	return hash
}

// ---------------- PRIMITIVAS DEL DICCIONARIO ------------------------------------------
func (dic *hashCerrado[K, V]) Guardar(clave K, dato V) {
	// Si la clave ya existe, actualiza el dato o valor de esa clave
	fmt.Println(len(dic.tabla))
	if dic.Pertenece(clave) {
		id_hashing := dic.obtenerPosClave(clave)
		dic.tabla[id_hashing].valor = dato
		return
	}
	dic.cantidad++
	if dic.Cantidad() >= len(dic.tabla)/2 {
		dic.redimensionar(len(dic.tabla) * 2)
	}
	pos_clave := dic.f_hash(clave)
	// si la posición está vacia, guarda sin problemas
	if dic.tabla[pos_clave].estado == VACIO {
		celda := crearCelda(clave, dato)
		dic.tabla[pos_clave] = *celda
		dic.tabla[pos_clave].hashingID = pos_clave
		return
	}
	// si la posición está ocupada me fijo en las proximas posiciones respetando la constante
	pos_libre := dic.obtenerPosVacia(pos_clave)

	if pos_libre == -1 { // Si las K siguientes pos no estan vacias:
		//  Delego a las posiciones siguientes que encuentren un lugar vacío dentro del rango de Hopscotch
		var pos_sig int
		if pos_clave >= len(dic.tabla)-1 {
			pos_sig = 0
		} else {
			pos_sig = pos_clave + 1
		}
		pos_clave_siguiente := dic.tabla[pos_sig].hashingID
		pos_actual := pos_clave_siguiente
		for i := pos_clave_siguiente; i < CONSTANTE_HOPSCOTCH+pos_clave; i++ {
			pos_actual++
			pos_valida := dic.obtenerPosVacia(i)
			if pos_valida != -1 {
				pos_libre = pos_valida
				break
			}
		} //termina el ciclo

		// las siguientes posiciones no lograron encontrar celdas vacias, redimensiono
		if pos_libre == -1 {
			dic.redimensionar(len(dic.tabla) * 2)
			dic.cantidad--
			dic.Guardar(clave, dato)
			return
		}
		// la celda actual encontró una posicion vacía en sus siguientes posiciones respetando el rango
		dic.tabla[pos_actual], dic.tabla[pos_libre] = dic.tabla[pos_libre], dic.tabla[pos_actual]
		// ahora celda_actual es una celda VACIA
		celda := crearCelda(clave, dato)
		celda.hashingID = pos_clave
		dic.tabla[pos_actual] = *celda
	}
	// si en las K siguientes posiciones encuentré una vacía, guardo sin problemas
	celda := crearCelda(clave, dato)
	celda.hashingID = pos_clave
	dic.tabla[pos_libre] = *celda

}

func (dic *hashCerrado[K, V]) Pertenece(clave K) bool { //check
	pos_clave := dic.obtenerPosClave(clave)
	return pos_clave != -1 && dic.tabla[pos_clave].estado == OCUPADO
}

func (dic *hashCerrado[K, V]) Obtener(clave K) V { //check
	if !dic.Pertenece(clave) {
		panic("La clave no pertenece al diccionario")
	}
	pos_clave := dic.obtenerPosClave(clave)
	return dic.tabla[pos_clave].valor
}

func (dic *hashCerrado[K, V]) Borrar(clave K) V {
	pos_clave := dic.obtenerPosClave(clave)
	if pos_clave == -1 || dic.tabla[pos_clave].estado != OCUPADO {
		panic("La clave no pertenece al diccionario")
	}
	dic.cantidad--
	dato := dic.tabla[pos_clave].valor
	dic.tabla[pos_clave].estado = VACIO
	if dic.Cantidad() <= len(dic.tabla)/4 && len(dic.tabla) > 10 {
		dic.redimensionar(len(dic.tabla) / 2)
	}
	return dato
}

func (dic *hashCerrado[K, V]) Cantidad() int {
	return dic.cantidad
}

func (dic *hashCerrado[K, V]) Iterador() IterDiccionario[K, V] {
	iter := new(iteradorDiccionarioExterno[K, V])
	lista := TDALista.CrearListaEnlazada[Celda[V, K]]()
	for i, elem := range dic.tabla {
		if dic.tabla[i].estado == OCUPADO {
			lista.InsertarUltimo(elem)
		}
	}
	dic.iter = lista.Iterador()
	iter.iterador = &dic.iter
	return iter
}

func (dic *hashCerrado[K, V]) Iterar(visitar func(clave K, dato V) bool) {
	iter := dic.Iterador()
	for iter.HaySiguiente() {
		clave, dato := iter.VerActual()
		if !visitar(clave, dato) {
			return
		}
		iter.Siguiente()
	}

}

// -----------------PRIMITIVAS DEL ITERADOR ---------------------------

func (iter *iteradorDiccionarioExterno[K, V]) HaySiguiente() bool {
	return (*iter.iterador).HaySiguiente()

}

func (iter *iteradorDiccionarioExterno[K, V]) Siguiente() K {
	actual := (*iter.iterador).Siguiente()
	return actual.clave
}

func (iter *iteradorDiccionarioExterno[K, V]) VerActual() (K, V) {
	actual := (*iter.iterador).VerActual()
	return actual.clave, actual.valor
}

// --------------------------------------------------------------------

// Encuentra una posición válida, de no encontrarse retorna -1
func (dic *hashCerrado[K, V]) obtenerPosVacia(pos int) int {
	i := pos + 1
	for contador := 1; contador < CONSTANTE_HOPSCOTCH; contador++ {
		if i == len(dic.tabla) {
			i = 0
		}
		if dic.tabla[i].estado == VACIO {
			return i
		}
		i++
	}
	return -1
}

func (dic *hashCerrado[K, V]) redimensionar(capacidad int) {
	nuevo_dic := make([]Celda[V, K], capacidad)
	iter := dic.Iterador()
	dic.tabla = nuevo_dic
	for iter.HaySiguiente() {
		clave, valor := iter.VerActual()
		nuevo_id_hashing := dic.f_hash(clave)
		nueva_celda := crearCelda(clave, valor)
		nuevo_dic[nuevo_id_hashing] = *nueva_celda
		nuevo_dic[nuevo_id_hashing].hashingID = nuevo_id_hashing
		iter.Siguiente()
	}
}

func (dic *hashCerrado[K, V]) obtenerPosClave(clave K) int {
	pos_clave := dic.f_hash(clave)
	if dic.tabla[pos_clave].clave == clave {
		return pos_clave
	}

	i := pos_clave + 1
	for contador := 1; contador < CONSTANTE_HOPSCOTCH; contador++ {
		if i == len(dic.tabla) {
			i = 0
		}
		if dic.tabla[i].clave == clave {
			return i
		}
	}

	return -1
}

// ----------FUNCIONES PARA HASHING ---------------------------------

func convertirABytes[K comparable](clave K) []byte {
	return []byte(fmt.Sprintf("%v", clave))
}

func sdbmHash(data []byte) int64 {
	var hash int64

	for _, b := range data {
		hash = int64(b) + (hash << 6) + (hash << 16) - hash
	}

	return hash
}

func (dic *hashCerrado[K, V]) f_hash(clave K) int {
	clave_bytes := convertirABytes(clave)
	id_hashing := sdbmHash(clave_bytes)
	return modulo(int(id_hashing)) % len(dic.tabla)
}

func modulo(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}
