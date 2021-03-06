package services

import (
    "appengine"
    "appengine/urlfetch"
    "encoding/json"
    "io/ioutil"
    "math/rand"
    "net/http"
    "strings"
    "time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Return url response
func getResponse(context appengine.Context, url string) string {
    client := urlfetch.Client(context)
    resp, err := client.Get(url)
    if err != nil {
        context.Infof("GetBoards client.Get: %v", err.Error())
        return ""
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        context.Infof("GetBoards ioutil.ReadAll: %v", err.Error())
        return ""
    }
    return string(body)
}

// Return event type and description to post.
func GetEventData(request *http.Request) (string, string) {
    hookType := getHookType(request)
    var decoder *json.Decoder
    if hookType == "travis" {
        payload := request.FormValue("payload")
        decoder = json.NewDecoder(strings.NewReader(payload))
    } else if hookType == "pingdom" {
        message := request.FormValue("message")
        decoder = json.NewDecoder(strings.NewReader(message))
    } else {
        decoder = json.NewDecoder(request.Body)
    }
    switch hookType {
    case "github":
        return getGithubData(
            decoder, request.Header.Get("X-Github-Event"))
    case "doorbell":
        return getDoorbellData(decoder)
    case "bitbucket":
        return getBitbucketData(decoder, request.Header.Get("X-Event-Key"))
    case "travis":
        return getTravisData(decoder)
    case "teamcity":
        return getTeamcityData(decoder)
    case "pingdom":
        return getPingdomData(decoder)
    }
    return "", ""
}

// Return type of hook.
func getHookType(request *http.Request) string {
    if request.Header.Get("X-Github-Event") != "" {
        return "github"
    } else if request.Header.Get("X-Sender") == "Doorbell" {
        return "doorbell"
    } else if strings.Index(request.Header.Get("User-Agent"), "Bitbucket") > -1 {
        return "bitbucket"
    } else if request.Header.Get("Travis-Repo-Slug") != "" {
        return "travis"
    } else if strings.Index(request.Header.Get("User-Agent"), "Jakarta") > -1 {
        return "teamcity"
    } else if request.FormValue("message") != "" {
        return "pingdom"
    }
    return ""
}

// Return random alphanumeric string
func GetAlphaNumberic(n int) string {
    var src = rand.NewSource(time.Now().UnixNano())
    b := make([]byte, n)
    // A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
    for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = src.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }
    return string(b)
}
