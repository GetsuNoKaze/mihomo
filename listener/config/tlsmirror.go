package config

import "github.com/metacubex/mihomo/transport/tlsmirror"

type TLSMirrorConfig struct {
	PrimaryKey                    string                         `yaml:"primary-key" json:"primary-key,omitempty"`
	Dest                          string                         `yaml:"dest" json:"dest,omitempty"`
	Proxy                         string                         `yaml:"proxy" json:"proxy,omitempty"`
	ExplicitNonceCipherSuites     []uint16                       `yaml:"explicit-nonce-ciphersuites" json:"explicit-nonce-ciphersuites,omitempty"`
	DeferInstanceDerivedWriteTime TLSMirrorTimeSpec              `yaml:"defer-instance-derived-write-time" json:"defer-instance-derived-write-time,omitempty"`
	TransportLayerPadding         TLSMirrorTransportLayerPadding `yaml:"transport-layer-padding" json:"transport-layer-padding,omitempty"`
	ConnectionEnrolment           *TLSMirrorConnectionEnrolment  `yaml:"connection-enrolment" json:"connection-enrolment,omitempty"`
	SequenceWatermarkingEnabled   bool                           `yaml:"sequence-watermarking-enabled" json:"sequence-watermarking-enabled,omitempty"`
}

type TLSMirrorConnectionEnrolment struct {
	PrimaryIngressOutbound string `yaml:"primary-ingress-outbound" json:"primary-ingress-outbound,omitempty"`
	PrimaryEgressOutbound  string `yaml:"primary-egress-outbound" json:"primary-egress-outbound,omitempty"`
}

func (e *TLSMirrorConnectionEnrolment) Build() *tlsmirror.ConnectionEnrolment {
	if e == nil {
		return nil
	}
	return &tlsmirror.ConnectionEnrolment{
		PrimaryIngressOutbound: e.PrimaryIngressOutbound,
		PrimaryEgressOutbound:  e.PrimaryEgressOutbound,
	}
}

type TLSMirrorTimeSpec struct {
	BaseNanoseconds                    uint64 `yaml:"base-nanoseconds" json:"base-nanoseconds,omitempty"`
	UniformRandomMultiplierNanoseconds uint64 `yaml:"uniform-random-multiplier-nanoseconds" json:"uniform-random-multiplier-nanoseconds,omitempty"`
}

func (s TLSMirrorTimeSpec) Build() tlsmirror.TimeSpec {
	return tlsmirror.TimeSpec{
		BaseNanoseconds:                    s.BaseNanoseconds,
		UniformRandomMultiplierNanoseconds: s.UniformRandomMultiplierNanoseconds,
	}
}

type TLSMirrorTransportLayerPadding struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty"`
}

func (p TLSMirrorTransportLayerPadding) Build() tlsmirror.TransportLayerPadding {
	return tlsmirror.TransportLayerPadding{
		Enabled: p.Enabled,
	}
}
