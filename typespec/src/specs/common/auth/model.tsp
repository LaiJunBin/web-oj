namespace Common.AuthSpec;

model Token {
    accessToken: string;
    refreshToken: string;
}

model Profile {
    name: string;
    username: string;
    roles: string[];
}
