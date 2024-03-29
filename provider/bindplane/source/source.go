package source

import (
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/BlueMedoraPublic/terraform-provider-bindplane/provider/bindplane/common"

	"github.com/BlueMedoraPublic/bpcli/bindplane/sdk"
	"github.com/BlueMedoraPublic/bpcli/util/uuid"
)

const testConnectionERR = "test connection error"

// Result describes the response from the bindplane source
// create api
type Result struct {
	SourceID string
	JobID    string
}

// Create attempts to create a source repeatedly until it is
// created succefully or timeout is exceeded
func Create(source sdk.SourceConfigCreate, timeout int) (Result, error) {
	r := Result{}

	if err := source.Validate(); err != nil {
		return r, errors.Wrap(err, "Not attempting to create source, validation failed")
	}

	config, err := buildConfig(source)
	if err != nil {
		return r, errors.Wrap(err, "Not attempting to create source, buildConfig() failed")
	}

	bp, err := common.New()
	if err != nil {
		return r, errors.Wrap(err, "Not attempting to create source, failed to run sdk.New()")
	}

	startTime := time.Now().Unix()

	for {
		resp, err := bp.CreateSource(config)
		if err != nil {
			return r, errors.Wrap(err, string(config))
		}
		r.JobID = resp.JobID

		// monitor the job until it finish
		r.SourceID, err = watchJob(r.JobID)

		// if there is an error, try again until timeout reached
		if err != nil {
			timeCurrent := time.Now().Unix()
			if (timeCurrent - startTime) > int64(timeout) {
				return r, errors.Wrap(err, "Timeout exceeded for source creation. JobID: "+r.JobID)
			}

			time.Sleep(5 * time.Second)
			continue

		}

		return r, nil
	}
}

// Read returns a source with a given ID
func Read(id string) (sdk.SourceConfigGet, error) {
	bp, err := common.New()
	if err != nil {
		return sdk.SourceConfigGet{}, err
	}
	return bp.GetSource(id)
}

// Delete deletes a source
func Delete(id string) error {
	bp, err := common.New()
	if err != nil {
		return err
	}

	body, err := bp.DeleteSource(id)
	if err != nil {
		return errors.Wrap(err, string(body))
	}
	return nil
}

// returns the sourceID from a completed job
func watchJob(jobID string) (string, error) {
	bp, err := common.New()
	if err != nil {
		return "", errors.Wrap(err, "Not attempting to monitor job "+jobID)
	}

	for {
		job, err := bp.GetJob(jobID)
		if err != nil {
			return "", errors.Wrap(err, "Failed to call sdk.GetJob() with job id "+jobID)
		}

		complete, err := parseStatus(job)
		if err != nil {
			return "", err
		}

		if complete == true {
			return getSourceID(bp, jobID)
		}

		time.Sleep(5 * time.Second)
		continue
	}
}

// ParseStatus returns true if job complete and an error
// if job status failed or unexpected
func parseStatus(job sdk.Job) (bool, error) {
	status := strings.ToLower(job.Status)
	if status == "complete" {
		return true, nil
	} else if status == "in progress" {
		return false, nil
	} else if status == "testing connection to source" {
		return false, nil
	} else if status == "queued for completion" {
		return false, nil
	} else if status == "failed" {

		jobMsg := strings.ToLower(job.Message)
		subStr := "test connection failed"
		if strings.Contains(jobMsg, subStr) == true {
			return false, errors.New(testConnectionERR)
		}
		return false, errors.Wrap(jobErr(job), "job: "+job.ID+" failed. "+job.Message)

	}
	return false, errors.Wrap(jobErr(job), "ParseStatus() failed to parse job id "+job.ID)
}

/*
getSourceID returns a source uuid from a job. This is not safe
to call unless you know the job was a source create api call,
and the source has been created succefully.
*/
func getSourceID(bp *sdk.BindPlane, jobID string) (string, error) {
	j, err := bp.GetJob(jobID)
	if err != nil {
		return "", errors.Wrap(err, "Attempted to get source ID from Job with ID: "+jobID)
	}

	// not safe, could fail if the source was not created
	id := j.Result.(map[string]interface{})["id"].(string)

	if uuid.IsUUID(id) == false {
		err := errors.New("uuid.IsUUID returned false when checking job id " + jobID)
		err = errors.Wrap(err, "getSourceID() may have been given a bad job id")
		return "", err

	}
	return id, nil
}

func buildConfig(source sdk.SourceConfigCreate) ([]byte, error) {
	x, err := json.Marshal(source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert source struct to []byte with json.Marshal")
	}
	return x, err
}

func jobErr(job sdk.Job) error {
	s := job.Status
	m := job.Message
	r, ok := job.Result.(string)
	if ok != true {
		return errors.New("status: " + s + " message: " + m)
	}
	return errors.New("status: " + s + " message: " + m + " result: " + r)
}
