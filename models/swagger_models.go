package models

// SwaggerUser hanya untuk dokumentasi swagger
type SwaggerUser struct {
    ID        uint    `json:"id"`
    CreatedAt string  `json:"created_at"`
    UpdatedAt string  `json:"updated_at"`
    DeletedAt *string `json:"deleted_at"`

    Name      string  `json:"name"`
    Email     string  `json:"email"`
    Role      string  `json:"role"`
}

// SwaggerKebun hanya untuk swagger
type SwaggerKebun struct {
    ID        uint    `json:"id"`
    CreatedAt string  `json:"created_at"`
    UpdatedAt string  `json:"updated_at"`
    DeletedAt *string `json:"deleted_at"`

    Nama      string  `json:"nama"`
    Lokasi    string  `json:"lokasi"`
}

// SwaggerTanaman hanya untuk swagger
type SwaggerTanaman struct {
    ID        uint          `json:"id"`
    CreatedAt string        `json:"created_at"`
    UpdatedAt string        `json:"updated_at"`
    DeletedAt *string       `json:"deleted_at"`

    Nama      string        `json:"nama"`
    Harga     int           `json:"harga"`

    KebunID   uint          `json:"kebun_id"`
    Kebun     SwaggerKebun  `json:"kebun"`
}

// SwaggerBooking hanya untuk swagger
type SwaggerBooking struct {
    ID        uint            `json:"id"`
    CreatedAt string          `json:"created_at"`
    UpdatedAt string          `json:"updated_at"`
    DeletedAt *string         `json:"deleted_at"`

    UserID    uint            `json:"user_id"`
    User      SwaggerUser     `json:"user"`

    TanamanID uint            `json:"tanaman_id"`
    Tanaman   SwaggerTanaman  `json:"tanaman"`
}