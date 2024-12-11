package main

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

type Product struct {
	ID    int
	Name  string
	Price float64
}

var (
	catalog = []Product{
		{ID: 1, Name: "Товар 1", Price: 100.0},
		{ID: 2, Name: "Товар 2", Price: 200.0},
	}
	cart = []Product{}
	mu   sync.Mutex
)

var templates = template.Must(template.ParseFiles(
	"templates/catalog.html",
	"templates/cart.html",
	"templates/add_product.html",
))

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	data := struct {
		Catalog []Product
		Cart    []Product
	}{
		Catalog: catalog,
		Cart:    cart,
	}

	templates.ExecuteTemplate(w, "catalog.html", data)
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Некорректный ID товара", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for _, product := range catalog {
		if product.ID == id {
			cart = append(cart, product)
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

//

func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Некорректный ID товара", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, product := range catalog {
		if product.ID == id {
			catalog = append(catalog[:i], catalog[i+1:]...)
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func addProductFormHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "add_product.html", nil)
}

func addProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	priceStr := r.FormValue("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Некорректная цена", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	newProduct := Product{
		ID:    len(catalog) + 1,
		Name:  name,
		Price: price,
	}
	catalog = append(catalog, newProduct)

	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	http.HandleFunc("/", catalogHandler)
	http.HandleFunc("/add-to-cart", addToCartHandler)
	http.HandleFunc("/delete-product", deleteProductHandler)
	http.HandleFunc("/add-product", addProductFormHandler)
	http.HandleFunc("/add-product-submit", addProductHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8181", nil)
}