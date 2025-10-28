package database

import (
	"fmt"
	"math/rand"
	"time"
)

// SeedDatabase peuple toutes les tables de la base de données
func SeedDatabase(years int) error {
	fmt.Println("🌱 Génération des données de référence...")

	// 1. Générer les fournisseurs
	supplierIDs, err := seedSuppliers(20)
	if err != nil {
		return fmt.Errorf("erreur génération fournisseurs: %w", err)
	}

	// 2. Générer les produits
	productIDs, err := seedProducts(100, supplierIDs)
	if err != nil {
		return fmt.Errorf("erreur génération produits: %w", err)
	}

	// 3. Lier produits et catégories (les catégories sont déjà créées dans init.sql)
	categoryIDs, err := getCategoryIDs()
	if err != nil {
		return fmt.Errorf("erreur récupération catégories: %w", err)
	}

	err = seedProductCategories(productIDs, categoryIDs)
	if err != nil {
		return fmt.Errorf("erreur liaison produits-catégories: %w", err)
	}

	// 4. Générer les clients
	customerIDs, err := seedCustomers(1000)
	if err != nil {
		return fmt.Errorf("erreur génération clients: %w", err)
	}

	// 5. Générer les magasins
	storeIDs, err := seedStores(10)
	if err != nil {
		return fmt.Errorf("erreur génération magasins: %w", err)
	}

	// 6. Récupérer les méthodes de paiement (déjà créées dans init.sql)
	paymentMethodIDs, err := getPaymentMethodIDs()
	if err != nil {
		return fmt.Errorf("erreur récupération méthodes paiement: %w", err)
	}

	// 7. Générer les promotions
	promotionIDs, err := seedPromotions(15)
	if err != nil {
		return fmt.Errorf("erreur génération promotions: %w", err)
	}

	// 8. Générer les commandes et lignes de commande
	fmt.Println("🌱 Génération des commandes et ventes...")
	err = seedOrdersAndItems(years, customerIDs, storeIDs, paymentMethodIDs, promotionIDs, productIDs)
	if err != nil {
		return fmt.Errorf("erreur génération commandes: %w", err)
	}

	// 9. Analyse finale
	fmt.Println("🔍 Analyse des tables...")
	_, err = DB.Exec("ANALYZE")
	if err != nil {
		fmt.Println("⚠️ Attention: échec de l'analyse:", err)
	}

	return nil
}

// seedSuppliers génère les fournisseurs
func seedSuppliers(count int) ([]int, error) {
	fmt.Printf("   📦 Génération de %d fournisseurs...\n", count)

	supplierNames := []string{
		"TechSupply Co", "ElectroWorld", "FashionHub", "FoodDistrib", "HomeStyle",
		"SportGear Inc", "BookWorks", "ToyFactory", "BeautyPro", "GlobalImport",
		"MegaSupply", "PrimeVendor", "QualityGoods", "FastShip", "DirectSource",
		"TopQuality", "ValueSupply", "BestChoice", "SmartDistrib", "ProSupplier",
	}

	cities := []string{"Paris", "Lyon", "Marseille", "Toulouse", "Bordeaux", "Nice", "Nantes", "Strasbourg"}
	countries := []string{"France", "Allemagne", "Italie", "Espagne", "Belgique"}

	ids := make([]int, 0, count)

	for i := 0; i < count; i++ {
		name := supplierNames[i%len(supplierNames)]
		if i >= len(supplierNames) {
			name = fmt.Sprintf("%s %d", name, i/len(supplierNames)+1)
		}

		var id int
		err := DB.QueryRow(`
			INSERT INTO suppliers (name, contact_name, email, phone, city, country)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, name,
			fmt.Sprintf("Contact %d", i+1),
			fmt.Sprintf("contact%d@%s.com", i+1, "supplier"),
			fmt.Sprintf("01%08d", rand.Intn(100000000)),
			cities[rand.Intn(len(cities))],
			countries[rand.Intn(len(countries))],
		).Scan(&id)

		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	fmt.Printf("   ✅ %d fournisseurs créés\n", len(ids))
	return ids, nil
}

// seedProducts génère les produits
func seedProducts(count int, supplierIDs []int) ([]int, error) {
	fmt.Printf("   📦 Génération de %d produits...\n", count)

	productPrefixes := []string{
		"Smartphone", "Laptop", "Tablet", "TV", "Camera",
		"T-shirt", "Jeans", "Sneakers", "Jacket", "Dress",
		"Pasta", "Rice", "Coffee", "Tea", "Juice",
		"Sofa", "Chair", "Table", "Lamp", "Rug",
		"Ball", "Racket", "Bike", "Weights", "Yoga Mat",
	}

	ids := make([]int, 0, count)

	for i := 0; i < count; i++ {
		prefix := productPrefixes[rand.Intn(len(productPrefixes))]
		name := fmt.Sprintf("%s %d", prefix, i+1)
		price := 10.0 + rand.Float64()*490.0
		stock := rand.Intn(1000)
		supplierID := supplierIDs[rand.Intn(len(supplierIDs))]

		var id int
		err := DB.QueryRow(`
			INSERT INTO products (name, description, supplier_id, base_price, stock_quantity)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, name,
			fmt.Sprintf("Description du produit %s", name),
			supplierID,
			price,
			stock,
		).Scan(&id)

		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	fmt.Printf("   ✅ %d produits créés\n", len(ids))
	return ids, nil
}

// getCategoryIDs récupère les IDs des catégories
func getCategoryIDs() ([]int, error) {
	rows, err := DB.Query("SELECT id FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// seedProductCategories lie les produits aux catégories
func seedProductCategories(productIDs, categoryIDs []int) error {
	fmt.Printf("   🔗 Liaison produits-catégories...\n")

	count := 0
	for _, productID := range productIDs {
		// Chaque produit a 1 à 3 catégories
		numCategories := 1 + rand.Intn(3)
		usedCategories := make(map[int]bool)

		for i := 0; i < numCategories; i++ {
			categoryID := categoryIDs[rand.Intn(len(categoryIDs))]

			// Éviter les doublons
			if usedCategories[categoryID] {
				continue
			}
			usedCategories[categoryID] = true

			_, err := DB.Exec(`
				INSERT INTO product_categories (product_id, category_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, productID, categoryID)

			if err != nil {
				return err
			}
			count++
		}
	}

	fmt.Printf("   ✅ %d liaisons créées\n", count)
	return nil
}

// seedCustomers génère les clients
func seedCustomers(count int) ([]int, error) {
	fmt.Printf("   👥 Génération de %d clients...\n", count)

	firstNames := []string{"Jean", "Marie", "Pierre", "Sophie", "Luc", "Anne", "Paul", "Julie", "Marc", "Claire"}
	lastNames := []string{"Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit", "Durand", "Leroy", "Moreau"}
	cities := []string{"Paris", "Lyon", "Marseille", "Toulouse", "Bordeaux", "Nice", "Nantes", "Strasbourg", "Montpellier", "Lille"}

	ids := make([]int, 0, count)

	for i := 0; i < count; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		email := fmt.Sprintf("%s.%s%d@email.com", firstName, lastName, i+1)
		city := cities[rand.Intn(len(cities))]

		var id int
		err := DB.QueryRow(`
			INSERT INTO customers (first_name, last_name, email, phone, city, postal_code, country)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`, firstName, lastName, email,
			fmt.Sprintf("06%08d", rand.Intn(100000000)),
			city,
			fmt.Sprintf("%05d", 10000+rand.Intn(90000)),
			"France",
		).Scan(&id)

		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	fmt.Printf("   ✅ %d clients créés\n", len(ids))
	return ids, nil
}

// seedStores génère les magasins
func seedStores(count int) ([]int, error) {
	fmt.Printf("   🏪 Génération de %d magasins...\n", count)

	stores := []struct {
		name   string
		city   string
		region string
	}{
		{"Store Paris Centre", "Paris", "Île-de-France"},
		{"Store Lyon Part-Dieu", "Lyon", "Auvergne-Rhône-Alpes"},
		{"Store Marseille Vieux-Port", "Marseille", "Provence-Alpes-Côte d'Azur"},
		{"Store Toulouse Capitole", "Toulouse", "Occitanie"},
		{"Store Bordeaux Chartrons", "Bordeaux", "Nouvelle-Aquitaine"},
		{"Store Nice Promenade", "Nice", "Provence-Alpes-Côte d'Azur"},
		{"Store Nantes Commerce", "Nantes", "Pays de la Loire"},
		{"Store Strasbourg Centre", "Strasbourg", "Grand Est"},
		{"Store Lille Europe", "Lille", "Hauts-de-France"},
		{"Store Montpellier Odysseum", "Montpellier", "Occitanie"},
	}

	ids := make([]int, 0, count)

	for i := 0; i < count && i < len(stores); i++ {
		store := stores[i]

		var id int
		err := DB.QueryRow(`
			INSERT INTO stores (name, city, region, country)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, store.name, store.city, store.region, "France").Scan(&id)

		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	fmt.Printf("   ✅ %d magasins créés\n", len(ids))
	return ids, nil
}

// getPaymentMethodIDs récupère les IDs des méthodes de paiement
func getPaymentMethodIDs() ([]int, error) {
	rows, err := DB.Query("SELECT id FROM payment_methods WHERE active = true ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// seedPromotions génère les promotions
func seedPromotions(count int) ([]int, error) {
	fmt.Printf("   🎁 Génération de %d promotions...\n", count)

	promoNames := []string{
		"Soldes d'été", "Black Friday", "Cyber Monday", "Noël", "Nouvel An",
		"Printemps", "Rentrée", "Saint-Valentin", "Pâques", "Fête des Mères",
		"Fête des Pères", "Halloween", "Anniversaire magasin", "Vente privée", "Flash sale",
	}

	ids := make([]int, 0, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		code := fmt.Sprintf("PROMO%d", i+1)
		name := promoNames[i%len(promoNames)]
		discount := float64(5 + rand.Intn(46)) // 5% à 50%

		// Dates aléatoires dans le passé
		daysAgo := rand.Intn(365 * 2)
		startDate := now.AddDate(0, 0, -daysAgo)
		endDate := startDate.AddDate(0, 0, 7+rand.Intn(23)) // 7 à 30 jours

		var id int
		err := DB.QueryRow(`
			INSERT INTO promotions (code, name, discount_percent, start_date, end_date, active)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, code, name, discount, startDate, endDate, rand.Float32() > 0.3).Scan(&id)

		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	fmt.Printf("   ✅ %d promotions créées\n", len(ids))
	return ids, nil
}

// seedOrdersAndItems génère les commandes et lignes de commande
func seedOrdersAndItems(years int, customerIDs, storeIDs, paymentMethodIDs, promotionIDs, productIDs []int) error {
	totalDays := years * 365
	totalOrders := 0
	totalItems := 0

	startTime := time.Now()

	for day := 0; day < totalDays; day++ {
		orderDate := time.Now().AddDate(0, 0, -day)

		// 20 à 100 commandes par jour
		numOrders := 20 + rand.Intn(81)

		for i := 0; i < numOrders; i++ {
			// Créer une commande
			customerID := customerIDs[rand.Intn(len(customerIDs))]
			storeID := storeIDs[rand.Intn(len(storeIDs))]
			paymentMethodID := paymentMethodIDs[rand.Intn(len(paymentMethodIDs))]

			// 30% de chance d'avoir une promotion
			var promotionID *int
			if rand.Float32() < 0.3 && len(promotionIDs) > 0 {
				promID := promotionIDs[rand.Intn(len(promotionIDs))]
				promotionID = &promID
			}

			// Créer la commande (on calculera le total après)
			var orderID int64
			err := DB.QueryRow(`
				INSERT INTO orders (customer_id, store_id, payment_method_id, promotion_id, order_date, total_amount, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, customerID, storeID, paymentMethodID, promotionID, orderDate, 0, "completed").Scan(&orderID)

			if err != nil {
				return err
			}

			// Ajouter 1 à 5 produits dans cette commande
			numItems := 1 + rand.Intn(5)
			orderTotal := 0.0

			for j := 0; j < numItems; j++ {
				productID := productIDs[rand.Intn(len(productIDs))]
				quantity := 1 + rand.Intn(5)

				// Récupérer le prix du produit
				var basePrice float64
				err := DB.QueryRow("SELECT base_price FROM products WHERE id = $1", productID).Scan(&basePrice)
				if err != nil {
					return err
				}

				// Petite variation de prix (+/- 10%)
				unitPrice := basePrice * (0.9 + rand.Float64()*0.2)
				subtotal := unitPrice * float64(quantity)
				orderTotal += subtotal

				// Insérer la ligne de commande
				_, err = DB.Exec(`
					INSERT INTO order_items (order_id, product_id, quantity, unit_price, subtotal)
					VALUES ($1, $2, $3, $4, $5)
				`, orderID, productID, quantity, unitPrice, subtotal)

				if err != nil {
					return err
				}

				totalItems++
			}

			// Mettre à jour le total de la commande
			_, err = DB.Exec("UPDATE orders SET total_amount = $1 WHERE id = $2", orderTotal, orderID)
			if err != nil {
				return err
			}

			totalOrders++
		}

		if (day+1)%100 == 0 {
			fmt.Printf("   ... %d jours traités (%d commandes, %d lignes)\n", day+1, totalOrders, totalItems)
		}
	}

	fmt.Printf("   ✅ %d commandes créées avec %d lignes en %v\n", totalOrders, totalItems, time.Since(startTime))
	return nil
}
