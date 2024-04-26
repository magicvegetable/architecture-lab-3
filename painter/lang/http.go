package lang

import (
	"io"
	"net/http"
	"strings"

	"github.com/magicvegetable/architecture-lab-3/painter"
	"sync"
)

func HttpHandler(loop *painter.Loop, p *Parser) http.Handler {

	var parserM, posterM sync.Mutex
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var in io.Reader = r.Body

		if r.Method == http.MethodGet {
			in = strings.NewReader(r.URL.Query().Get("cmd"))
		}

		parserM.Lock()

		events, err := p.ParseOperations(in)

		if err != nil {
			// TODO: handle errors
			println(err)
			rw.WriteHeader(http.StatusBadRequest)
			parserM.Unlock()
			return
		}

		if len(events) != 0 {
			go func() {
				posterM.Lock()
				parserM.Unlock()

				loop.PostEvents(events)

				posterM.Unlock()
			}()
		} else {
			parserM.Unlock()
		}

		rw.WriteHeader(http.StatusOK)
	})
}
