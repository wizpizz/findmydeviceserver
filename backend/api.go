package backend

import "net/http"
import "os"
import "github.com/rs/zerolog/log"

var remoteIpHeaderName string = ""

func getRemoteIp(r *http.Request) string {
	remoteIp := r.Header.Get(remoteIpHeaderName)
	if remoteIp == "" {
		remoteIp = r.RemoteAddr
	}
	return remoteIp
}

// Adds various security headers.
// Check your deployment with https://securityheaders.com.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Xss-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self' ; img-src 'self' data: https://*.tile.openstreetmap.org ; script-src 'self' 'wasm-unsafe-eval' ; upgrade-insecure-requests")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=()")
		w.Header().Set("Referrer-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

func buildServeMux(webDir string, config config) *http.ServeMux {
	// Workaround: cache value in global field to avoid needing to pass down the config into the API code
	remoteIpHeaderName = config.RemoteIpHeader

	var envToken string
	log.Info().Msgf("config.RegistrationToken: %s", config.RegistrationToken)
	if config.RegistrationToken == "env" {
		log.Info().Msg("RegistrationToken is set to env")
		envToken = os.Getenv("REGISTRATION_TOKEN")
		if envToken != "" {
			log.Info().Msgf("envToken: %s", envToken)
			config.RegistrationToken = envToken
		} else {
			log.Info().Msg("REGISTRATION_TOKEN is not set")
		}
	} else {
		log.Info().Msgf("RegistrationToken is set to %s", config.RegistrationToken)
		envToken = config.RegistrationToken
	}

	mainDeviceHandler := mainDeviceHandler{createDeviceHandler{envToken}}

	apiV1Mux := http.NewServeMux()
	apiV1Mux.HandleFunc("/command", mainCommand)
	apiV1Mux.HandleFunc("/command/", mainCommand)
	//Disabled Feature: CommandLogs
	//apiV1Mux.HandleFunc("/commandLogs", getCommandLog)
	//apiV1Mux.HandleFunc("/commandLogs/", getCommandLog)
	apiV1Mux.HandleFunc("/location", mainLocation)
	apiV1Mux.HandleFunc("/location/", mainLocation)
	apiV1Mux.HandleFunc("/locations", getAllLocations)
	apiV1Mux.HandleFunc("/locations/", getAllLocations)
	apiV1Mux.HandleFunc("/locationDataSize", getLocationDataSize)
	apiV1Mux.HandleFunc("/locationDataSize/", getLocationDataSize)
	apiV1Mux.HandleFunc("/picture", mainPicture)
	apiV1Mux.HandleFunc("/picture/", mainPicture)
	apiV1Mux.HandleFunc("/pictures", getAllPictures)
	apiV1Mux.HandleFunc("/pictures/", getAllPictures)
	apiV1Mux.HandleFunc("/pictureSize", getPictureSize)
	apiV1Mux.HandleFunc("/pictureSize/", getPictureSize)
	apiV1Mux.HandleFunc("/key", getPrivKey)
	apiV1Mux.HandleFunc("/key/", getPrivKey)
	apiV1Mux.HandleFunc("/pubKey", getPubKey)
	apiV1Mux.HandleFunc("/pubKey/", getPubKey)
	apiV1Mux.Handle("/device", mainDeviceHandler)
	apiV1Mux.Handle("/device/", mainDeviceHandler)
	apiV1Mux.HandleFunc("/password", postPassword)
	apiV1Mux.HandleFunc("/password/", postPassword)
	apiV1Mux.HandleFunc("/push", mainPushUrl)
	apiV1Mux.HandleFunc("/push/", mainPushUrl)
	apiV1Mux.HandleFunc("/salt", requestSalt)
	apiV1Mux.HandleFunc("/salt/", requestSalt)
	apiV1Mux.HandleFunc("/requestAccess", requestAccess)
	apiV1Mux.HandleFunc("/requestAccess/", requestAccess)
	apiV1Mux.HandleFunc("/version", getVersion)
	apiV1Mux.HandleFunc("/version/", getVersion)

	// Uncomment this once the API v1 is no longer hosted at the root "/" (because we cannot have two "/" in muxFinal).
	// Until then, as a side-effect, the static files are also served under /api/v1/.
	// staticFilesMux := http.NewServeMux()
	// staticFilesMux.Handle("/", http.FileServer(http.Dir(webDir)))
	apiV1Mux.Handle("/", http.FileServer(http.Dir(webDir)))

	muxFinal := http.NewServeMux()
	// muxFinal.Handle("/", securityHeadersMiddleware(staticFilesMux))
	muxFinal.Handle("/", securityHeadersMiddleware(apiV1Mux)) // deprecated
	muxFinal.Handle("/api/v1/", http.StripPrefix("/api/v1", securityHeadersMiddleware(apiV1Mux)))

	return muxFinal
}
