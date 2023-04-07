package view

import (
	"embed"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

//go:embed ui
var f embed.FS

type ViewConfig struct {
}

func (p *ViewConfig) OnEvent(event any) {

}

var _ = InstallPlugin(&ViewConfig{})

func (p *ViewConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/view/" {
		var s string
		Streams.Range(func(streamPath string, _ *Stream) {
			s += fmt.Sprintf("<a href='%s'>%s</a><br>", streamPath, streamPath)
		})
		if s != "" {
			s = "<b>Live Streams</b><br>" + s
		}
		for name, p := range Plugins {
			if pullcfg, ok := p.Config.(config.PullConfig); ok {
				if pullonsub := pullcfg.GetPullConfig().PullOnSub; pullonsub != nil {
					s += fmt.Sprintf("<b>%s pull stream on subscribe</b><br>", name)
					for streamPath, url := range pullonsub {
						s += fmt.Sprintf("<a href='%s'>%s</a> <-- %s<br>", streamPath, streamPath, url)
					}
				}
			}
		}
		w.Write([]byte(s))
		return
	}
	ss := strings.Split(r.URL.Path, "/")
	if b, err := f.ReadFile("ui/" + ss[len(ss)-1]); err == nil {
		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(ss[len(ss)-1])))
		w.Write(b)
	} else {
		//w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		//w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		b, err = f.ReadFile("ui/index.html")
		w.Write(b)
	}
}

type ViewStream struct {
	Source     string
	StreamPath string
}

func filterStreams() (ss []*ViewStream) {

	//优先获取配置文件中视频流
	for name, p := range Plugins {
		if pullcfg, ok := p.Config.(config.PullConfig); ok {

			if pullonstart := pullcfg.GetPullConfig().PullOnStart; pullonstart != nil {
				//s += fmt.Sprintf("<b>%s pull stream on subscribe</b><br>", name)
				//for streamPath, url := range pullonsub {
				var sourcename string = name
				sourcename += "Pull"

				for streamPath := range pullonstart {
					ss = append(ss, &ViewStream{sourcename, streamPath})
					//s += fmt.Sprintf("<a href='%s'>%s</a> <-- %s<br>", streamPath, streamPath, url)
				}
			}

			if pullonsub := pullcfg.GetPullConfig().PullOnSub; pullonsub != nil {
				//s += fmt.Sprintf("<b>%s pull stream on subscribe</b><br>", name)
				//for streamPath, url := range pullonsub {
				var sourcename string = name

				sourcename += "Pull"

				for streamPath := range pullonsub {
					ss = append(ss, &ViewStream{sourcename, streamPath})
					//s += fmt.Sprintf("<a href='%s'>%s</a> <-- %s<br>", streamPath, streamPath, url)
				}
			}
		}
	}

	//过滤出动态添加的视频流
	Streams.RLock()
	defer Streams.RUnlock()

	Streams.Range(func(streamPath string, s *Stream) {
		//s += fmt.Sprintf("<a href='%s'>%s</a><br>", streamPath, streamPath)
		var isrepeat bool = false
		for _, s := range ss {
			if streamPath == s.StreamPath {
				isrepeat = true
			}
		}
		if !isrepeat {
			ss = append(ss, &ViewStream{"api", streamPath})
		}
	})

	return
}

func (*ViewConfig) API_streamslist(w http.ResponseWriter, r *http.Request) {
	util.ReturnJson(filterStreams, time.Second, w, r)
}
