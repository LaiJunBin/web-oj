using Common.AuthSpec;

namespace Admin.UserSpec;

model User {
    @key id: int32;
    ...Profile;
    password?: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}
