package coincount

var (
	VisaCard = Account{
		ID:   1020,
		Name: "Visa Card",
	}

	GeminiUSD = Account{
		ID:   1021,
		Name: "Gemini USD",
	}

	EthMain = Account{
		ID:   1330,
		Name: "ETH-Main",
	}

	EthCoinbase = Account{
		ID:   1031,
		Name: "ETH-Coinbase",
	}

	EthGemini = Account{
		ID:   1032,
		Name: "ETH-Gemini",
	}

	EnsDomains = Account{
		ID:   1510,
		Name: "ENS Domains",
	}

	ElectricBill = Account{
		ID:   2350,
		Name: "Electric Bill",
	}

	RevenueEth = Account{
		ID:   4010,
		Name: "Revenue ETH",
	}

	CostOfEthSold = Account{
		ID:   5010,
		Name: "Cost of ETH Sold",
	}

	EthAdjustments = Account{
		ID:   5800,
		Name: "Eth Adjustments",
	}

	EthTXFee = Account{
		ID:   6200,
		Name: "Ethereum Transaction Fee",
	}

	CoinbaseFee = Account{
		ID:   6201,
		Name: "Coinbase Fee",
	}

	GeminiFee = Account{
		ID:   6202,
		Name: "Gemini Fee",
	}

	AssetSales = Account{
		ID:   7900,
		Name: "Gain/Loss Asset Sales",
	}

	GLAccounts = []Account{
		VisaCard,
		GeminiUSD,
		EthMain,
		EthCoinbase,
		EthGemini,
		EnsDomains,
		ElectricBill,
		RevenueEth,
		CostOfEthSold,
		EthAdjustments,
		EthTXFee,
		CoinbaseFee,
		GeminiFee,
		AssetSales,
	}

	Ether = Item{
		ID:   1,
		Name: "Ether",
	}

	ExpenseItem = Item{
		ID:   999999,
		Name: "Generic Expense",
	}

	InventoryItems = []Item{
		Ether,
		ExpenseItem,
	}

	ElectricCompany = Vendor{
		ID:   1,
		Name: "Electric Company",
	}

	ENSRegistrar = Vendor{
		ID:   3,
		Name: "ENS Registrar",
	}

	Coinbase = Vendor{
		ID:   4,
		Name: "Coinbase",
	}

	Gemini = Vendor{
		ID:   5,
		Name: "Gemini",
	}

	Vendors = []Vendor{
		ElectricCompany,
		ENSRegistrar,
		Coinbase,
		Gemini,
	}
)
