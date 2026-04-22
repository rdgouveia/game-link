package store

import (
	"context"
	"errors"
	"fmt"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

const errCodeUniqueViolation = "23505"

var (
	ErrNotFound    = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Storer interface {
	NewUser(ctx context.Context, username, email, password, fullname string) (*model.User, error)
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userID, fullname, biography string) error
	UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error
	NewGame(ctx context.Context, name, summary, imageUrl string, igdbID int32) (*model.Game, error)
	GetGameByIgdbID(ctx context.Context, igdbID int32) (*model.Game, error)
	NewUserGame(ctx context.Context, userID, gameID string) error
	GetUserGames(ctx context.Context, userID string) ([]model.Game, error)
	RemoveUserGame(ctx context.Context, userID, gameID string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	SearchUsersByFullname(ctx context.Context, name, currentUserID string) ([]model.UserSearchResult, error)
	GetUserByUsername(ctx context.Context, username, currentUserID string) (*model.UserSearchResult, error)
	GetMatchCandidates(ctx context.Context, userID string) ([]model.MatchCandidate, error)
	AvailabilityStorer
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == errCodeUniqueViolation
}

func (s *Store) NewUser(ctx context.Context, username, email, password, fullname string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	row, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Email:    email,
		Password: string(hash),
		Fullname: fullname,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	return &model.User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
		Fullname: row.Fullname,
		Verified: row.Verified,
	}, nil
}

func (s *Store) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &model.User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
		Fullname: row.Fullname,
		Verified: row.Verified,
	}, nil
}

func (s *Store) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := s.q.GetUserByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &model.User{
		ID:        uuidToString(row.ID),
		Username:  row.Username,
		Email:     row.Email,
		Fullname:  row.Fullname,
		Biography: row.Biography.String,
		Verified:  row.Verified,
		AvatarURL: row.AvatarUrl.String,
	}, nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, userID, fullname, biography string) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return err
	}
	return s.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		Fullname:  fullname,
		Biography: pgtype.Text{String: biography, Valid: biography != ""},
		ID:        userUUID,
	})
}

func (s *Store) NewGame(ctx context.Context, name, summary, imageUrl string, igdbID int32) (*model.Game, error) {
	row, err := s.q.CreateGame(ctx, db.CreateGameParams{
		Name:     name,
		Summary:  summary,
		ImageUrl: imageUrl,
		IgdbID:   igdbID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return &model.Game{
		ID:       uuidToString(row.ID),
		Name:     row.Name,
		Summary:  row.Summary,
		ImageUrl: row.ImageUrl,
		IgdbID:   row.IgdbID,
	}, nil
}

func (s *Store) GetGameByIgdbID(ctx context.Context, igdbID int32) (*model.Game, error) {
	row, err := s.q.GetGameByIgdbID(ctx, igdbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &model.Game{
		ID:       uuidToString(row.ID),
		Name:     row.Name,
		Summary:  row.Summary,
		ImageUrl: row.ImageUrl,
		IgdbID:   row.IgdbID,
	}, nil
}

func (s *Store) NewUserGame(ctx context.Context, userID, gameID string) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	gameUUID, err := stringToUUID(gameID)
	if err != nil {
		return fmt.Errorf("invalid game ID: %w", err)
	}

	return s.q.AddGameToUser(ctx, db.AddGameToUserParams{
		UserID: userUUID,
		GameID: gameUUID,
	})
}

func (s *Store) GetUserGames(ctx context.Context, userID string) ([]model.Game, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	rows, err := s.q.GetGamesByUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	games := make([]model.Game, len(rows))
	for i, row := range rows {
		games[i] = model.Game{
			ID:       uuidToString(row.ID),
			Name:     row.Name,
			Summary:  row.Summary,
			ImageUrl: row.ImageUrl,
			IgdbID:   row.IgdbID,
		}
	}
	return games, nil
}

func (s *Store) SearchUsersByFullname(ctx context.Context, name, currentUserID string) ([]model.UserSearchResult, error) {
	currentUUID, err := stringToUUID(currentUserID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.SearchUsersByFullname(ctx, db.SearchUsersByFullnameParams{
		Column1:     pgtype.Text{String: name, Valid: true},
		RequesterID: currentUUID,
	})
	if err != nil {
		return nil, err
	}
	results := make([]model.UserSearchResult, len(rows))
	for i, row := range rows {
		results[i] = model.UserSearchResult{
			ID:         uuidToString(row.ID),
			Username:   row.Username,
			Fullname:   row.Fullname,
			Biography:  row.Biography.String,
			Verified:   row.Verified,
			AvatarURL:  row.AvatarUrl.String,
			LinkStatus: resolveLinkStatus(row.LinkStatus, row.LinkRequesterID, currentUserID),
		}
	}
	return results, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username, currentUserID string) (*model.UserSearchResult, error) {
	currentUUID, err := stringToUUID(currentUserID)
	if err != nil {
		return nil, err
	}
	row, err := s.q.GetUserByUsername(ctx, db.GetUserByUsernameParams{
		Username:    username,
		RequesterID: currentUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &model.UserSearchResult{
		ID:         uuidToString(row.ID),
		Username:   row.Username,
		Fullname:   row.Fullname,
		Biography:  row.Biography.String,
		Verified:   row.Verified,
		AvatarURL:  row.AvatarUrl.String,
		LinkStatus: resolveLinkStatus(row.LinkStatus, row.LinkRequesterID, currentUserID),
	}, nil
}

// resolveLinkStatus converte os dados brutos do banco no status legível pelo front.
func resolveLinkStatus(status pgtype.Text, requesterID pgtype.UUID, currentUserID string) string {
	if !status.Valid {
		return "none"
	}
	if status.String == "accepted" {
		return "accepted"
	}
	if uuidToString(requesterID) == currentUserID {
		return "pending_sent"
	}
	return "pending_received"
}

func (s *Store) UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return err
	}
	return s.q.UpdateUserAvatar(ctx, db.UpdateUserAvatarParams{
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: true},
		ID:        userUUID,
	})
}

func (s *Store) EmailExists(ctx context.Context, email string) (bool, error) {
	return s.q.CheckEmailExists(ctx, email)
}

func (s *Store) UsernameExists(ctx context.Context, username string) (bool, error) {
	return s.q.CheckUsernameExists(ctx, username)
}

func (s *Store) RemoveUserGame(ctx context.Context, userID, gameID string) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	gameUUID, err := stringToUUID(gameID)
	if err != nil {
		return fmt.Errorf("invalid game ID: %w", err)
	}

	return s.q.RemoveGameFromUser(ctx, db.RemoveGameFromUserParams{
		UserID: userUUID,
		GameID: gameUUID,
	})
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func stringToUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}
