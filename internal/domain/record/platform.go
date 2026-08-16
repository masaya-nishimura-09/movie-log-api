package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Platform string

const (
	PlatformMubi             Platform = "mubi"
	PlatformNetflix          Platform = "netflix"
	PlatformAmazonPrimeVideo Platform = "amazon_prime_video"
	PlatformDisneyPlus       Platform = "disney_plus"
	PlatformAppleTVPlus      Platform = "apple_tv_plus"

	PlatformMax              Platform = "max"
	PlatformPeacock          Platform = "peacock"
	PlatformParamountPlus    Platform = "paramount_plus"
	PlatformCriterionChannel Platform = "criterion_channel"
	PlatformShudder          Platform = "shudder"
	PlatformTubi             Platform = "tubi"
	PlatformPlutoTV          Platform = "pluto_tv"
	PlatformRokuChannel      Platform = "roku_channel"
	PlatformVRV              Platform = "vrv"

	PlatformBBCIPlayer Platform = "bbc_iplayer"
	PlatformChannel4   Platform = "channel_4"
	PlatformITVX       Platform = "itvx"

	PlatformUNext         Platform = "u_next"
	PlatformHulu          Platform = "hulu"
	PlatformABEMA         Platform = "abema"
	PlatformDAnimeStore   Platform = "d_anime_store"
	PlatformLemino        Platform = "lemino"
	PlatformFOD           Platform = "fod"
	PlatformParavi        Platform = "paravi"
	PlatformWOWOWOnDemand Platform = "wowow_on_demand"

	PlatformTheater     Platform = "theater"
	PlatformDVDBluRay   Platform = "dvd_bluray"
	PlatformTVBroadcast Platform = "tv_broadcast"
	PlatformRental      Platform = "rental"

	PlatformOther Platform = "other"
)

func NewPlatform(value string) (Platform, error) {
	if value == "" {
		return "", fmt.Errorf("%w: platform is required", exception.ErrInvalid)
	}

	switch platform := Platform(value); platform {
	case PlatformMubi,
		PlatformNetflix,
		PlatformAmazonPrimeVideo,
		PlatformDisneyPlus,
		PlatformAppleTVPlus,
		PlatformMax,
		PlatformPeacock,
		PlatformParamountPlus,
		PlatformCriterionChannel,
		PlatformShudder,
		PlatformTubi,
		PlatformPlutoTV,
		PlatformRokuChannel,
		PlatformVRV,
		PlatformBBCIPlayer,
		PlatformChannel4,
		PlatformITVX,
		PlatformUNext,
		PlatformHulu,
		PlatformABEMA,
		PlatformDAnimeStore,
		PlatformLemino,
		PlatformFOD,
		PlatformParavi,
		PlatformWOWOWOnDemand,
		PlatformTheater,
		PlatformDVDBluRay,
		PlatformTVBroadcast,
		PlatformRental,
		PlatformOther:
		return platform, nil
	default:
		return "", fmt.Errorf("%w: invalid platform", exception.ErrInvalid)
	}
}
