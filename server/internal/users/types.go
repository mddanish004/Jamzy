package users

type PublicUser struct {
	ID          string `json:"id"`
	ClerkUserID string `json:"clerkUserId"`
	Username    string `json:"username"`
}

type MeResponse struct {
	Authenticated      bool        `json:"authenticated"`
	User               *PublicUser `json:"user"`
	OnboardingComplete bool        `json:"onboardingComplete"`
}

type ProfileResponse struct {
	User PublicUser `json:"user"`
}

type CreateProfileBody struct {
	Username any `json:"username"`
}
