package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var (
	users = map[string]string{
		"user1": "password123",
	}
	balances = map[string]float64{
		"user1": 100,
	}
	products = map[string]float64{
		"apple":  1.0,
		"banana": 0.5,
	}
	cart = make(map[string]map[string]int)
	mu   = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add_to_cart/", addToCartHandler)
	http.HandleFunc("/checkout", checkoutHandler)
	http.ListenAndServe(":9090", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if storedPass, ok := users[username]; ok && storedPass == password {
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: username,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	//Menambah body karena tidak bisa input username dan password
	fmt.Fprint(w, `<body><form action="/login" method="post">
Username: <input type="text" name="username"><br>
Password: <input type="password" name="password"><br>
<input type="submit" value="Login">
</form></body>`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}
	//menambah body karena tidak bisa menampilkan product di etalase
	/*
			= float64(24.12)
		var str = strconv.FormatFloat(num, 'f', 6, 64)

		fmt.Println(str) // 24.120000
	*/
	// \n ganti ke <br>
	fmt.Fprintf(w, "<body> Hello, %s! Available products:<br>", username)

	var str = strconv.FormatFloat(balances[username], 'f', 2, 64)
	fmt.Fprintf(w, "Your Balance: %s<br>", str)

	for product, price := range products {
		fmt.Fprintf(w, "%s: $%.2f <a href=\"/add_to_cart/%s\">Add to cart</a><br>", product, price, product)
	}

	fmt.Fprint(w, "<br>Your cart: <br>")
	for product, quantity := range userCart {
		fmt.Fprintf(w, "%s: %d<br>", product, quantity)
	}

	fmt.Fprint(w, `<br><a href="/checkout">Checkout</a></body>`)
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	product := r.URL.Path[len("/add_to_cart/"):] //Sign
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}

	userCart[product]++ //sign
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	total := 0.0
	for product, quantity := range userCart {
		price, ok := products[product]
		/*
			"apple":  1.0,
			"banana": 0.5,

			products[apple]
		*/
		if !ok {
			continue
		}
		total += price * float64(quantity)
		/*
			total = total + price * quantity
			1 = 1 + 1.0 * 1
			2 = 2 + 0.5 * 2
			3
		*/
	}
	balances[username] -= total

	cart[username] = make(map[string]int) // reset keranjang setelah checkout
	/*
		Balances[user1] = Balance(100) - Total
		Balances[user1] = Balance(100) - 3
		Balances[user1] = Balance(97) - Total
	*/

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
