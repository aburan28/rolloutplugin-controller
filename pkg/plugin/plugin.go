package plugin

type Plugin struct {
	
}

type RolloutPlugin interface {
	SetWeight()
	SetMirrorRoute()
	Rollback()
	SetCanaryScale()
}
