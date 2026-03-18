package ai

// MCCEntry описывает одну категорию транзакции и её влияние на RPG-атрибуты персонажа.
// MCC (Merchant Category Code) — четырёхзначный код ISO 18245, который платёжные сети
// присваивают каждому мерчанту. Это детерминированный слой нашей гибридной классификации:
// если MCC есть в таблице, точность 100%. LLM используется только как fallback.
type MCCEntry struct {
	Category string
	// RPG-атрибуты: положительное значение — прибавляем, отрицательное — вычитаем
	XP        int
	HP        int
	Mana      int
	Strength  int
	Intellect int
	Luck      int
}

// mccTable — главная таблица классификации.
// Покрывает 50+ реальных категорий трат.
// Ключ — точный MCC-код, значение — категория + RPG-дельта.
var mccTable = map[int]MCCEntry{
	// ───────────────────────────── ПРОДУКТЫ И ЕДА ─────────────────────────────
	5411: {Category: "Supermarkets & Groceries", XP: 2, HP: 3},
	5412: {Category: "Grocery Stores", XP: 2, HP: 2},
	5422: {Category: "Freezer & Locker Meat", XP: 1, HP: 1},
	5441: {Category: "Candy & Confections", XP: 1, HP: -1},
	5451: {Category: "Dairy Products", XP: 1, HP: 2},
	5462: {Category: "Bakeries", XP: 1, HP: 1},
	5499: {Category: "Specialty Food Stores", XP: 2, HP: 2},

	// ───────────────────────────── РЕСТОРАНЫ ─────────────────────────────────
	5811: {Category: "Caterers", XP: 2, HP: -1, Mana: 1},
	5812: {Category: "Restaurants & Cafes", XP: 2, HP: -2, Mana: 2},
	5813: {Category: "Bars & Taverns", XP: 1, HP: -2, Mana: 3},
	5814: {Category: "Fast Food", XP: 1, HP: -3, Mana: 1},
	5815: {Category: "Digital Goods", XP: 3, Intellect: 1},

	// ───────────────────────────── СПОРТ И ФИТНЕС ────────────────────────────
	5912: {Category: "Drug Stores & Pharmacies", XP: 2, HP: 3},
	5940: {Category: "Sporting Goods", XP: 3, Strength: 2},
	5941: {Category: "Sports & Outdoor Shops", XP: 3, Strength: 3},
	5945: {Category: "Hobby, Toy & Game Shops", XP: 2, Mana: 2},
	7011: {Category: "Hotels & Lodging", XP: 5, Mana: 3},
	7012: {Category: "Timeshares", XP: 3, Mana: 2},
	7941: {Category: "Sports Clubs & Stadiums", XP: 4, Strength: 3, Mana: 2},
	7991: {Category: "Amusement Parks", XP: 3, Mana: 4},
	7992: {Category: "Golf Courses", XP: 4, Strength: 2, Luck: 1},
	7993: {Category: "Video Game Arcades", XP: 2, Mana: 3},
	7996: {Category: "Amusement Attractions", XP: 2, Mana: 3},
	7997: {Category: "Country Clubs & Fitness", XP: 4, Strength: 4},
	7998: {Category: "Aquariums & Zoos", XP: 2, Mana: 2},
	7999: {Category: "Recreation Services", XP: 3, Strength: 2},

	// ───────────────────────────── ОБРАЗОВАНИЕ ───────────────────────────────
	5942: {Category: "Book Stores", XP: 5, Intellect: 4},
	5943: {Category: "Stationery & Supplies", XP: 2, Intellect: 1},
	8211: {Category: "Schools & Education", XP: 8, Intellect: 6},
	8220: {Category: "Colleges & Universities", XP: 10, Intellect: 8},
	8241: {Category: "Correspondence Schools", XP: 6, Intellect: 5},
	8244: {Category: "Business & Secretarial Schools", XP: 5, Intellect: 4},
	8249: {Category: "Vocational Schools", XP: 5, Intellect: 4},
	8299: {Category: "Educational Services", XP: 6, Intellect: 5},

	// ───────────────────────────── МЕДИЦИНА И ЗДОРОВЬЕ ───────────────────────
	8011: {Category: "Doctors & Physicians", XP: 3, HP: 5},
	8021: {Category: "Dentists", XP: 3, HP: 4},
	8031: {Category: "Osteopaths", XP: 3, HP: 4},
	8041: {Category: "Chiropractors", XP: 3, HP: 3},
	8049: {Category: "Podiatrists", XP: 2, HP: 3},
	8050: {Category: "Nursing & Care Facilities", XP: 2, HP: 3},
	8062: {Category: "Hospitals", XP: 2, HP: 6},
	8099: {Category: "Healthcare Services", XP: 3, HP: 4},
	5122: {Category: "Drugs & Pharmaceuticals", XP: 2, HP: 5},

	// ───────────────────────────── ТРАНСПОРТ ─────────────────────────────────
	4111: {Category: "Local Transit & Commuter", XP: 2},
	4112: {Category: "Passenger Railways", XP: 3, Mana: 1},
	4121: {Category: "Taxicabs & Limousines", XP: 2},
	4131: {Category: "Bus Lines", XP: 1},
	4411: {Category: "Cruise Lines", XP: 8, Mana: 6},
	4511: {Category: "Airlines", XP: 10, Mana: 5},
	4722: {Category: "Travel Agencies", XP: 6, Mana: 4},
	4784: {Category: "Tolls & Bridge Fees", XP: 1},
	7512: {Category: "Car Rentals", XP: 4, Strength: 1},
	5541: {Category: "Service Stations & Gas", XP: 2},
	5542: {Category: "Automated Fuel Dispensers", XP: 1},
	5571: {Category: "Motorcycle Shops", XP: 3, Strength: 2},

	// ───────────────────────────── ОДЕЖДА И ШОППИНГ ──────────────────────────
	5611: {Category: "Men's Clothing Stores", XP: 3, Mana: 1},
	5621: {Category: "Women's Clothing Stores", XP: 3, Mana: 1},
	5631: {Category: "Women's Accessories", XP: 2, Mana: 1},
	5641: {Category: "Children's Clothing", XP: 2},
	5651: {Category: "Family Clothing Stores", XP: 2},
	5661: {Category: "Shoe Stores", XP: 2, Mana: 1},
	5691: {Category: "Men's & Women's Clothing", XP: 3, Mana: 1},
	5699: {Category: "Apparel & Accessories", XP: 2, Mana: 1},
	5944: {Category: "Jewelry & Watches", XP: 4, Luck: 3},
	5947: {Category: "Gift & Novelty Shops", XP: 2, Mana: 2},

	// ───────────────────────────── ЭЛЕКТРОНИКА ───────────────────────────────
	5045: {Category: "Computers & Peripherals", XP: 5, Intellect: 3},
	5065: {Category: "Electronics Parts", XP: 3, Intellect: 2},
	5732: {Category: "Electronics Stores", XP: 4, Intellect: 3},
	5734: {Category: "Computer & Software Stores", XP: 5, Intellect: 4},
	5735: {Category: "Record & Music Stores", XP: 3, Mana: 3},

	// ───────────────────────────── РАЗВЛЕЧЕНИЯ ───────────────────────────────
	7832: {Category: "Movie Theaters", XP: 3, Mana: 4},
	7922: {Category: "Theatrical Producers", XP: 4, Mana: 5, Intellect: 1},
	7929: {Category: "Bands & Orchestras", XP: 4, Mana: 5},
	7933: {Category: "Bowling Alleys", XP: 2, Strength: 1, Mana: 2},
	5816: {Category: "Digital Games", XP: 2, Mana: 4},

	// ───────────────────────────── АЗАРТНЫЕ ИГРЫ ─────────────────────────────
	7995: {Category: "Gambling & Casinos", XP: 1, Luck: 5, Mana: -1},

	// ───────────────────────────── ФИНАНСЫ ───────────────────────────────────
	6011: {Category: "ATM Cash Withdrawal", XP: 0},
	6012: {Category: "Financial Institutions", XP: 1},
	6051: {Category: "Quasi-Cash & Currency", XP: 0},
	6300: {Category: "Insurance", XP: 2, HP: 1},
	6411: {Category: "Insurance Premiums", XP: 2, HP: 1},

	// ───────────────────────────── ПОДПИСКИ И СЕРВИСЫ ────────────────────────
	4814: {Category: "Telecom Services", XP: 2, Intellect: 1},
	4816: {Category: "Computer Network Services", XP: 3, Intellect: 2},
	4899: {Category: "Cable & Pay TV", XP: 1, Mana: 2},
	4900: {Category: "Utilities (Electric, Gas)", XP: 1},

	// ───────────────────────────── ДОМ И РЕМОНТ ──────────────────────────────
	5200: {Category: "Home Supply & Hardware", XP: 3, Strength: 2},
	5211: {Category: "Lumber & Building Materials", XP: 2, Strength: 2},
	5231: {Category: "Glass & Paint", XP: 2, Strength: 1},
	5251: {Category: "Hardware Stores", XP: 2, Strength: 2},
	5261: {Category: "Lawn & Garden", XP: 2, Strength: 1},
	5712: {Category: "Furniture Stores", XP: 3},
	5713: {Category: "Floor Covering Stores", XP: 2},
	5714: {Category: "Drapery & Upholstery", XP: 1},
	5719: {Category: "Misc Home Furnishing", XP: 2},
	5722: {Category: "Household Appliance Stores", XP: 3},

	// ───────────────────────────── ЛИЧНЫЙ УХОД ───────────────────────────────
	7230: {Category: "Beauty Salons & Barbershops", XP: 2, Mana: 2},
	7297: {Category: "Massage & Health Spas", XP: 3, HP: 3, Mana: 2},
	7298: {Category: "Health & Beauty Spas", XP: 3, HP: 2, Mana: 2},
	5977: {Category: "Cosmetics & Beauty", XP: 2, Mana: 1},

	// ───────────────────────────── ЕДА ДОСТАВКА ──────────────────────────────
	5963: {Category: "Direct Selling (Food)", XP: 2, HP: 1},
	5992: {Category: "Florists", XP: 2, Mana: 2},
	5999: {Category: "Misc Retail Stores", XP: 1},
}

// LookupMCC возвращает категорию и RPG-импакт для заданного MCC-кода.
// Возвращает (entry, true) если MCC найден, иначе (zero, false).
// Вызывается первым в цепочке классификации — до обращения к LLM.
func LookupMCC(mcc int) (MCCEntry, bool) {
	if mcc <= 0 {
		return MCCEntry{}, false
	}
	entry, ok := mccTable[mcc]
	return entry, ok
}

// DefaultEntry — категория-заглушка, когда ни MCC, ни LLM не дали результата.
var DefaultEntry = MCCEntry{Category: "General Purchase", XP: 1}
