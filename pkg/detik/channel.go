package detik

type Channel struct {
	name    string
	baseURL string
}

func (c Channel) Name() string {
	return c.name
}

func (c Channel) BaseURL() string {
	return c.BaseURL()
}

var (
	ChannelNews = Channel{
		name:    "News",
		baseURL: "https://news.detik.com/indeks",
	}

	ChannelEdu = Channel{
		name:    "Edu",
		baseURL: "https://www.detik.com/edu/indeks",
	}

	ChannelFinance = Channel{
		name:    "Finance",
		baseURL: "https://finance.detik.com/indeks",
	}

	ChannelHot = Channel{
		name:    "Hot",
		baseURL: "https://hot.detik.com/indeks",
	}

	ChannelInet = Channel{
		name:    "Inet",
		baseURL: "https://inet.detik.com/indeks",
	}

	ChannelSport = Channel{
		name:    "Sport",
		baseURL: "https://sport.detik.com/indeks",
	}

	ChannelOto = Channel{
		name:    "Oto",
		baseURL: "https://oto.detik.com/indeks",
	}

	ChannelTravel = Channel{
		name:    "Travel",
		baseURL: "https://travel.detik.com/indeks",
	}

	ChannelSepakBola = Channel{
		name:    "Sepakbola",
		baseURL: "https://sport.detik.com/sepakbola/indeks",
	}

	ChannelFood = Channel{
		name:    "Food",
		baseURL: "https://food.detik.com/indeks",
	}

	ChannelHealth = Channel{
		name:    "Health",
		baseURL: "https://health.detik.com/indeks",
	}

	ChannelJatim = Channel{
		name:    "Jatim",
		baseURL: "https://www.detik.com/jatim/indeks",
	}

	ChannelJateng = Channel{
		name:    "Jateng",
		baseURL: "https://www.detik.com/jateng/indeks",
	}

	ChannelJabar = Channel{
		name:    "Jabar",
		baseURL: "https://www.detik.com/jabar/indeks",
	}

	ChannelSulsel = Channel{
		name:    "Sulsel",
		baseURL: "https://www.detik.com/sulsel/indeks",
	}

	ChannelSumut = Channel{
		name:    "Sumut",
		baseURL: "https://www.detik.com/sumut/indeks",
	}

	ChannelBali = Channel{
		name:    "Bali",
		baseURL: "https://www.detik.com/bali/indeks",
	}

	ChannelHikmah = Channel{
		name:    "Hikmah",
		baseURL: "https://www.detik.com/hikmah/indeks",
	}

	ChannelSumbagsel = Channel{
		name:    "Sumbagsel",
		baseURL: "https://www.detik.com/sumbagsel/indeks",
	}

	ChannelProperti = Channel{
		name:    "Properti",
		baseURL: "https://www.detik.com/sumbagsel/indeks",
	}

	ChannelJogja = Channel{
		name:    "Jogja",
		baseURL: "https://www.detik.com/jogja/indeks",
	}

	ChannelPop = Channel{
		name:    "Pop",
		baseURL: "https://www.detik.com/pop/indeks",
	}
)
