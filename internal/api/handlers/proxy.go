package handlers

import (
	"encoding/json"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"net/http"
	"os/exec"

	"github.com/jarm-proxy/internal/api/ctx"
	"github.com/jarm-proxy/internal/api/requests"
)

type JarmOutput struct {
	Host   string `json:"host"`
	Ip     string `json:"ip"`
	Result string `json:"result"`
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewProxyRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	//TODO: mv this stuff to use cases cause it can be separated for list of functions: GetJarmOutput, Compare, SendRequest etc.
	cmd := exec.Command("jarm.py", request.Address, "-j")
	out, err := cmd.Output()
	if err != nil {
		ctx.Log(r).WithError(err).Error("failed to execute cmd")
		ape.Render(w, problems.InternalError())
		return
	}

	var output JarmOutput
	err = json.Unmarshal(out, &output)
	if err != nil {
		ctx.Log(r).WithError(err).Error("failed to unmarshal jarm output")
		ape.Render(w, problems.InternalError())
		return
	}

	ctx.Log(r).WithFields(logan.F{
		"host": output.Host,
		"ip": output.Ip,
		"hash": output.Result,
	})

	fingerPrints := ctx.Cfg(r).Jarm().FingerPrints

	//TODO: refactor later
	for _, element := range fingerPrints {
		if output.Result == element {
			ctx.Log(r).Debug("should not be connected")
			ape.Render(w, problems.Forbidden())
			return
		}
	}

	//TODO: implement proxy stuff that will send request to address
	//TODO: mb refactor request etc.
}
