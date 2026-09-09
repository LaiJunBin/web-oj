package provider

type Config struct {
	Infrastructures []interface{}
	Controllers     []interface{}
	Presenters      []interface{}
	Services        []interface{}
	Usecases        []interface{}
	Applications    []interface{}
	Factories       []interface{}
	Handlers        []interface{}
	Providers       []interface{}
	Invokes         []interface{}
}

type Provider struct {
	Config Config
}

func New(config Config) *Provider {
	return &Provider{
		Config: config,
	}
}

func (p *Provider) AddProvider(provider *Provider) {
	p.Config.Infrastructures = append(p.Config.Infrastructures, provider.Config.Infrastructures...)
	p.Config.Controllers = append(p.Config.Controllers, provider.Config.Controllers...)
	p.Config.Presenters = append(p.Config.Presenters, provider.Config.Presenters...)
	p.Config.Services = append(p.Config.Services, provider.Config.Services...)
	p.Config.Usecases = append(p.Config.Usecases, provider.Config.Usecases...)
	p.Config.Applications = append(p.Config.Applications, provider.Config.Applications...)
	p.Config.Factories = append(p.Config.Factories, provider.Config.Factories...)
	p.Config.Handlers = append(p.Config.Handlers, provider.Config.Handlers...)
	p.Config.Providers = append(p.Config.Providers, provider.Config.Providers...)
	p.Config.Invokes = append(p.Config.Invokes, provider.Config.Invokes...)
}
