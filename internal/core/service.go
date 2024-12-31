package core

import "context"

type Service interface {
	HandleRequest(req SubmitRequest) error
}

type ServiceConfig struct {
	Storage         Storage
	VacancyResolver VacancyResolver
}

type service struct {
	ServiceConfig
}

func (s *service) HandleRequest(ctx context.Context, req SubmitRequest) error {
	// as per system specification, supposedly the request is processed asynchronously
	// however to make it simple, we will process it synchronously for now.
	var err error
	switch req.SubmissionType {
	case SubmitTypeManual:
		err = s.Storage.Save(ctx, req.Vacancy)
		if err != nil {
			return NewInternalError(err)
		}
	case SubmitTypeURL:
		// resolve apply url to get vacancy details
		vacancy, err := s.VacancyResolver.Resolve(ctx, req.Vacancy.ApplyURL)
		if err != nil {
			// the error will be determined by resolver implementation
			return err
		}

		// then save the job details
		err = s.Storage.Save(ctx, *vacancy)
		if err != nil {
			return NewInternalError(err)
		}
	default:
		return NewBadRequestError("Invalid submission type")
	}
	return nil
}
