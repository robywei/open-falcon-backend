package restful

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	commonNqmDb "github.com/Cepave/open-falcon-backend/common/db/nqm"
	commonGin "github.com/Cepave/open-falcon-backend/common/gin"
	"github.com/Cepave/open-falcon-backend/common/gin/mvc"
	commonModel "github.com/Cepave/open-falcon-backend/common/model"
	commonNqmModel "github.com/Cepave/open-falcon-backend/common/model/nqm"
	"github.com/spf13/cast"
)

func listPingtasks(
	c *gin.Context,
	q *commonNqmModel.PingtaskQuery,
	p *struct {
		Paging *commonModel.Paging `mvc:"pageSize[50] pageOrderBy[enable#desc:name#asc:num_of_enabled_agents#desc]"`
	},
) (*commonModel.Paging, mvc.OutputBody) {
	p.Paging = commonGin.PagingByHeader(c, p.Paging)
	pingtasks, resultPaging := commonNqmDb.ListPingtasks(q, *p.Paging)

	return resultPaging, mvc.JsonOutputBody(pingtasks)
}

func getPingtasksById(
	p *struct {
		PingtaskID int32 `mvc:"param[pingtask_id]"`
	},
) mvc.OutputBody {
	return mvc.JsonOutputOrNotFound(commonNqmDb.GetPingtaskById(p.PingtaskID))
}

func addNewPingtask(
	pm *commonNqmModel.PingtaskModify,
) mvc.OutputBody {
	pingtask := commonNqmDb.AddAndGetPingtask(pm)
	return mvc.JsonOutputBody2(http.StatusCreated, pingtask)
}

func modifyPingtask(
	p *struct {
		ID int32 `mvc:"param[pingtask_id]"`
	},
	pm *commonNqmModel.PingtaskModify,
) mvc.OutputBody {
	pingtask := commonNqmDb.UpdateAndGetPingtask(p.ID, pm)
	return mvc.JsonOutputBody(pingtask)
}

func addPingtaskToAgentForAgent(c *gin.Context) {
	/**
	 * Builds data from body of request
	 */
	var pingtaskIDStr string
	var pingtaskID int32

	if v, ok := c.GetQuery("pingtask_id"); ok {
		pingtaskIDStr = v
	}
	if v, err := cast.ToInt32E(pingtaskIDStr); err == nil {
		pingtaskID = v
	}

	var agentIDStr string
	var agentID int32
	if v := c.Param("agent_id"); v != "" {
		agentIDStr = v
	}
	if v, err := cast.ToInt32E(agentIDStr); err == nil {
		agentID = v
	}

	agentWithNewPingtask, err := commonNqmDb.AssignPingtaskToAgentForAgent(agentID, pingtaskID)
	if err != nil {
		switch err.(type) {
		case commonNqmDb.ErrDuplicatedNqmAgent:
			commonGin.JsonConflictHandler(
				c,
				commonGin.DataConflictError{
					ErrorCode:    1,
					ErrorMessage: err.Error(),
				},
			)
		default:
			panic(err)
		}

		return
	}

	c.JSON(http.StatusCreated, agentWithNewPingtask)
}

func removePingtaskFromAgentForAgent(c *gin.Context) {
	var agentIDStr string
	var agentID int32
	if v := c.Param("agent_id"); v != "" {
		agentIDStr = v
	}
	if v, err := cast.ToInt32E(agentIDStr); err == nil {
		agentID = v
	}

	var pingtaskIDStr string
	var pingtaskID int32
	if v := c.Param("pingtask_id"); v != "" {
		pingtaskIDStr = v
	}
	if v, err := cast.ToInt32E(pingtaskIDStr); err == nil {
		pingtaskID = v
	}

	agentWithRemovedPingtask, err := commonNqmDb.RemovePingtaskFromAgentForAgent(agentID, pingtaskID)
	if err != nil {
		switch err.(type) {
		case commonNqmDb.ErrDuplicatedNqmAgent:
			commonGin.JsonConflictHandler(
				c,
				commonGin.DataConflictError{
					ErrorCode:    1,
					ErrorMessage: err.Error(),
				},
			)
		default:
			panic(err)
		}

		return
	}
	c.JSON(http.StatusOK, agentWithRemovedPingtask)
}

func listTargetsOfAgent(c *gin.Context) {
	c.JSON(http.StatusCreated, "fuck you")
}

func addPingtaskToAgentForPingtask(c *gin.Context) {
	/**
	 * Builds data from body of request
	 */
	var pingtaskIDStr string
	var pingtaskID int32

	if v := c.Param("pingtask_id"); v != "" {
		pingtaskIDStr = v
	}
	if v, err := cast.ToInt32E(pingtaskIDStr); err == nil {
		pingtaskID = v
	}

	var agentIDStr string
	var agentID int32
	if v, ok := c.GetQuery("agent_id"); ok {
		agentIDStr = v
	}
	if v, err := cast.ToInt32E(agentIDStr); err == nil {
		agentID = v
	}

	agentWithNewPingtask, err := commonNqmDb.AssignPingtaskToAgentForPingtask(agentID, pingtaskID)
	if err != nil {
		switch err.(type) {
		case commonNqmDb.ErrDuplicatedNqmAgent:
			commonGin.JsonConflictHandler(
				c,
				commonGin.DataConflictError{
					ErrorCode:    1,
					ErrorMessage: err.Error(),
				},
			)
		default:
			panic(err)
		}

		return
	}

	c.JSON(http.StatusCreated, agentWithNewPingtask)
}

func removePingtaskFromAgentForPingtask(c *gin.Context) {
	var agentIDStr string
	var agentID int32
	if v := c.Param("agent_id"); v != "" {
		agentIDStr = v
	}
	if v, err := cast.ToInt32E(agentIDStr); err == nil {
		agentID = v
	}

	var pingtaskIDStr string
	var pingtaskID int32
	if v := c.Param("pingtask_id"); v != "" {
		pingtaskIDStr = v
	}
	if v, err := cast.ToInt32E(pingtaskIDStr); err == nil {
		pingtaskID = v
	}

	agentWithRemovedPingtask, err := commonNqmDb.RemovePingtaskFromAgentForPingtask(agentID, pingtaskID)
	if err != nil {
		switch err.(type) {
		case commonNqmDb.ErrDuplicatedNqmAgent:
			commonGin.JsonConflictHandler(
				c,
				commonGin.DataConflictError{
					ErrorCode:    1,
					ErrorMessage: err.Error(),
				},
			)
		default:
			panic(err)
		}

		return
	}
	c.JSON(http.StatusOK, agentWithRemovedPingtask)
}

func listAgentsByPingTask(
	query *commonNqmModel.AgentQueryWithPingTask,
	paging *struct {
		Page *commonModel.Paging `mvc:"pageSize[50] pageOrderBy[status#desc:connection_id#asc]"`
	},
) (*commonModel.Paging, mvc.OutputBody) {
	agents, resultPaging := commonNqmDb.ListAgentsWithPingTask(query, *paging.Page)

	return resultPaging, mvc.JsonOutputBody(agents)
}
