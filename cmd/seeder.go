// untuk menjalankan seeder nya
// go run -tags seeder ./config/seeder.go

//go:build seeder
// +build seeder

package main

import (
    "errors"
    "fmt"
    "log"
    "time"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "Avocycle/models"
    "Avocycle/config"
)

func main() {
    log.Println("[Seeder] mulai seeding data…")

    db, err := config.DbConnect()
    if err != nil {
        log.Fatalf("[Seeder] gagal konek DB: %v", err)
    }
    defer func() {
        sqlDB, err := db.DB()
        if err == nil {
            sqlDB.Close()
        }
    }()

    if err := runSeeder(db); err != nil {
        log.Fatalf("[Seeder] gagal seeding: %v", err)
    }

    log.Println("[Seeder] selesai ✅")
}

func runSeeder(db *gorm.DB) error {
    if err := db.AutoMigrate(
        &models.User{},
        &models.Kebun{},
        &models.Tanaman{},
        &models.Buah{},
        &models.FaseBunga{},
        &models.FaseBuah{},
        &models.FasePanen{},
        &models.ProsesProduksi{},
        &models.LogProsesProduksi{},
        &models.PenyakitTanaman{},
        &models.LogPenyakitTanaman{},
        &models.PerawatanPenyakit{},
        &models.Booking{},
    ); err != nil {
        return fmt.Errorf("auto migrate: %w", err)
    }

    users, err := seedUsers(db)
    if err != nil {
        return err
    }

    kebuns, err := seedKebuns(db)
    if err != nil {
        return err
    }

    tanamans, err := seedTanamans(db, kebuns)
    if err != nil {
        return err
    }

    prosesList, err := seedProsesProduksi(db)
    if err != nil {
        return err
    }

    buahs, err := seedBuahs(db, tanamans)
    if err != nil {
        return err
    }

    if err := seedFaseBunga(db, tanamans); err != nil {
        return err
    }
    if err := seedFaseBuah(db, tanamans); err != nil {
        return err
    }
    if err := seedFasePanen(db, tanamans); err != nil {
        return err
    }

    penyakits, err := seedPenyakit(db)
    if err != nil {
        return err
    }

    logPenyakit, err := seedLogPenyakit(db, tanamans, penyakits)
    if err != nil {
        return err
    }

    if err := seedPerawatan(db, logPenyakit); err != nil {
        return err
    }

    if err := seedLogProses(db, prosesList, buahs); err != nil {
        return err
    }

    if err := seedBookings(db, users, tanamans); err != nil {
        return err
    }

    return nil
}

// --- seed helpers ---

func seedUsers(db *gorm.DB) ([]models.User, error) {
    password := func(raw string) string {
        hash, _ := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
        return string(hash)
    }

    users := []models.User{
        {
            FullName:     "Admin Avocycle",
            Phone:        "081111111111",
            Email:        "admin@avocycle.test",
            PasswordHash: password("Admin123!"),
            AuthProvider: "Local",
            Role:         "Admin",
        },
        {
            FullName:     "Petani Andi",
            Phone:        "082222222222",
            Email:        "petani@avocycle.test",
            PasswordHash: password("Petani123!"),
            AuthProvider: "Local",
            Role:         "Petani",
        },
        {
            FullName:     "Pembeli Putri",
            Phone:        "083333333333",
            Email:        "pembeli@avocycle.test",
            PasswordHash: password("Pembeli123!"),
            AuthProvider: "Local",
            Role:         "Pembeli",
        },
    }

    for i := range users {
        if err := firstOrCreate(db, &users[i], "email = ?", users[i].Email); err != nil {
            return nil, fmt.Errorf("seed user %s: %w", users[i].Email, err)
        }
    }
    return users, nil
}

func seedKebuns(db *gorm.DB) ([]models.Kebun, error) {
    kebuns := []models.Kebun{
        {NamaKebun: "Kebun Sentra Alpukat", MDPL: "850"},
        {NamaKebun: "Kebun Percobaan Dataran Tinggi", MDPL: "1020"},
    }

    for i := range kebuns {
        if err := firstOrCreate(db, &kebuns[i], "nama_kebun = ?", kebuns[i].NamaKebun); err != nil {
            return nil, fmt.Errorf("seed kebun %s: %w", kebuns[i].NamaKebun, err)
        }
    }
    return kebuns, nil
}

func seedTanamans(db *gorm.DB, kebuns []models.Kebun) ([]models.Tanaman, error) {
    if len(kebuns) < 2 {
        return nil, errors.New("kebun belum tersedia cukup")
    }
    tanamans := []models.Tanaman{
        {
            NamaTanaman:  "Alpukat Mentega A1",
            Varietas:     "Var1",
            TanggalTanam: mustParseDate("2023-06-15"),
            KebunID:      kebuns[0].ID,
            KodeBlok:     "BLK-A",
            KodeTanaman:  "TAN-001",
            MasaProduksi: 6,
            FotoTanaman:  "https://res.cloudinary.com/demo/image/upload/v1/tanaman/alpukat-mentega-a1.jpg",
            FotoTanamanID:"tanaman/alpukat-mentega-a1",
        },
        {
            NamaTanaman:  "Alpukat Aligator B7",
            Varietas:     "Var2",
            TanggalTanam: mustParseDate("2023-09-05"),
            KebunID:      kebuns[0].ID,
            KodeBlok:     "BLK-B",
            KodeTanaman:  "TAN-002",
            MasaProduksi: 8,
        },
        {
            NamaTanaman:  "Alpukat Hass C3",
            Varietas:     "Var3",
            TanggalTanam: mustParseDate("2022-11-20"),
            KebunID:      kebuns[1].ID,
            KodeBlok:     "BLK-C",
            KodeTanaman:  "TAN-003",
            MasaProduksi: 10,
        },
    }

    for i := range tanamans {
        if err := firstOrCreate(db, &tanamans[i], "kode_tanaman = ?", tanamans[i].KodeTanaman); err != nil {
            return nil, fmt.Errorf("seed tanaman %s: %w", tanamans[i].KodeTanaman, err)
        }
    }
    return tanamans, nil
}

func seedBuahs(db *gorm.DB, tanamans []models.Tanaman) ([]models.Buah, error) {
    if len(tanamans) == 0 {
        return nil, errors.New("tanaman belum tersedia")
    }

    buahs := []models.Buah{
        {NamaBuah: "Panen Sesi 1", TanamanID: tanamans[0].ID},
        {NamaBuah: "Panen Sesi 2", TanamanID: tanamans[1].ID},
    }

    for i := range buahs {
        if err := firstOrCreate(db, &buahs[i], "nama_buah = ? AND tanaman_id = ?", buahs[i].NamaBuah, buahs[i].TanamanID); err != nil {
            return nil, fmt.Errorf("seed buah %s: %w", buahs[i].NamaBuah, err)
        }
    }
    return buahs, nil
}

func seedFaseBunga(db *gorm.DB, tanamans []models.Tanaman) error {
    if len(tanamans) == 0 {
        return nil
    }
    records := []models.FaseBunga{
        {
            MingguKe:     3,
            TanggalCatat: mustParseDatePtr("2023-10-03"),
            JumlahBunga:  56,
            BungaPecah:   12,
            PentilMuncul: 5,
            TanamanID:    tanamans[0].ID,
        },
        {
            MingguKe:     5,
            TanggalCatat: mustParseDatePtr("2023-11-02"),
            JumlahBunga:  72,
            BungaPecah:   18,
            PentilMuncul: 9,
            TanamanID:    tanamans[2].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i], "tanaman_id = ? AND minggu_ke = ?", records[i].TanamanID, records[i].MingguKe); err != nil {
            return fmt.Errorf("seed fase bunga tanaman %d minggu %d: %w", records[i].TanamanID, records[i].MingguKe, err)
        }
    }
    return nil
}

func seedFaseBuah(db *gorm.DB, tanamans []models.Tanaman) error {
    if len(tanamans) == 0 {
        return nil
    }
    records := []models.FaseBuah{
        {
            MingguKe:      8,
            TanggalCatat:  mustParseDatePtr("2023-12-12"),
            TanggalCover:  mustParseDatePtr("2023-12-05"),
            JumlahCover:   40,
            WarnaLabel:    "Kuning",
            EstimasiPanen: mustParseDatePtr("2024-01-20"),
            TanamanID:     tanamans[0].ID,
        },
        {
            MingguKe:      9,
            TanggalCatat:  mustParseDatePtr("2024-02-10"),
            TanggalCover:  mustParseDatePtr("2024-02-01"),
            JumlahCover:   35,
            WarnaLabel:    "Merah",
            EstimasiPanen: mustParseDatePtr("2024-03-05"),
            TanamanID:     tanamans[2].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i], "tanaman_id = ? AND minggu_ke = ?", records[i].TanamanID, records[i].MingguKe); err != nil {
            return fmt.Errorf("seed fase buah tanaman %d minggu %d: %w", records[i].TanamanID, records[i].MingguKe, err)
        }
    }
    return nil
}

func seedFasePanen(db *gorm.DB, tanamans []models.Tanaman) error {
    if len(tanamans) == 0 {
        return nil
    }
    records := []models.FasePanen{
        {
            TanggalPanenAktual: mustParseDatePtr("2024-02-01"),
            JumlahPanen:        320,
            JumlahSampel:       12,
            BeratTotal:         152.5,
            Catatan:            "Panen perdana dengan kualitas bagus.",
            TanamanID:          tanamans[0].ID,
        },
        {
            TanggalPanenAktual: mustParseDatePtr("2024-04-18"),
            JumlahPanen:        280,
            JumlahSampel:       10,
            BeratTotal:         133.0,
            Catatan:            "Panen kedua sedikit terhambat hujan.",
            TanamanID:          tanamans[1].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i], "tanaman_id = ? AND tanggal_panen_aktual = ?", records[i].TanamanID, records[i].TanggalPanenAktual); err != nil {
            return fmt.Errorf("seed fase panen tanaman %d: %w", records[i].TanamanID, err)
        }
    }
    return nil
}

func seedProsesProduksi(db *gorm.DB) ([]models.ProsesProduksi, error) {
    records := []models.ProsesProduksi{
        {Fase: "Berbunga"},
        {Fase: "Berbuah"},
        {Fase: "Panen"},
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i], "fase = ?", records[i].Fase); err != nil {
            return nil, fmt.Errorf("seed proses %s: %w", records[i].Fase, err)
        }
    }
    return records, nil
}

func seedLogProses(db *gorm.DB, proses []models.ProsesProduksi, buahs []models.Buah) error {
    if len(proses) < 3 || len(buahs) == 0 {
        return nil
    }

    records := []models.LogProsesProduksi{
        {
            Deskripsi:     "Pemangkasan ringan untuk merangsang bunga baru",
            PrediksiPanen: mustParseDate("2024-02-10"),
            ProsesID:      proses[0].ID,
            BuahID:        buahs[0].ID,
        },
        {
            Deskripsi:     "Pembungkusan buah untuk mencegah hama",
            PrediksiPanen: mustParseDate("2024-03-12"),
            ProsesID:      proses[1].ID,
            BuahID:        buahs[0].ID,
        },
        {
            Deskripsi:     "Seleksi buah matang dan penimbangan",
            PrediksiPanen: mustParseDate("2024-04-22"),
            ProsesID:      proses[2].ID,
            BuahID:        buahs[1].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i],
            "proses_id = ? AND buah_id = ? AND deskripsi = ?",
            records[i].ProsesID, records[i].BuahID, records[i].Deskripsi,
        ); err != nil {
            return fmt.Errorf("seed log proses %d: %w", i, err)
        }
    }
    return nil
}

func seedPenyakit(db *gorm.DB) ([]models.PenyakitTanaman, error) {
    records := []models.PenyakitTanaman{
        {NamaPenyakit: "Busuk Pangkal Batang", Deskripsi: "Disebabkan jamur, menyerang pangkal batang dan akar."},
        {NamaPenyakit: "Antraknosa", Deskripsi: "Cendawan Colletotrichum sp., menimbulkan bercak hitam pada buah."},
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i], "nama_penyakit = ?", records[i].NamaPenyakit); err != nil {
            return nil, fmt.Errorf("seed penyakit %s: %w", records[i].NamaPenyakit, err)
        }
    }
    return records, nil
}

func seedLogPenyakit(db *gorm.DB, tanamans []models.Tanaman, penyakits []models.PenyakitTanaman) ([]models.LogPenyakitTanaman, error) {
    if len(tanamans) == 0 || len(penyakits) == 0 {
        return nil, nil
    }

    records := []models.LogPenyakitTanaman{
        {
            Kondisi:    "Sedang",
            Catatan:    "Daun menguning dan terdapat bercak coklat.",
            Foto:       "https://res.cloudinary.com/demo/image/upload/v1/penyakit/busuk-pangkal.jpg",
            TanamanID:  tanamans[0].ID,
            PenyakitID: penyakits[0].ID,
        },
        {
            Kondisi:    "Ringan",
            Catatan:    "Ditemukan bercak hitam kecil pada buah muda.",
            TanamanID:  tanamans[1].ID,
            PenyakitID: penyakits[1].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i],
            "tanaman_id = ? AND penyakit_id = ? AND kondisi = ?",
            records[i].TanamanID, records[i].PenyakitID, records[i].Kondisi,
        ); err != nil {
            return nil, fmt.Errorf("seed log penyakit %d: %w", i, err)
        }
    }
    return records, nil
}

func seedPerawatan(db *gorm.DB, logs []models.LogPenyakitTanaman) error {
    if len(logs) == 0 {
        return nil
    }

    records := []models.PerawatanPenyakit{
        {
            Tindakan:             "Aplikasi fungisida sistemik setiap 7 hari.",
            LogPenyakitTanamanID: logs[0].ID,
        },
        {
            Tindakan:             "Penyemprotan fungisida kontak dosis rendah.",
            LogPenyakitTanamanID: logs[len(logs)-1].ID,
        },
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i],
            "log_penyakit_tanaman_id = ? AND tindakan = ?",
            records[i].LogPenyakitTanamanID, records[i].Tindakan,
        ); err != nil {
            return fmt.Errorf("seed perawatan %d: %w", i, err)
        }
    }
    return nil
}

func seedBookings(db *gorm.DB, users []models.User, tanamans []models.Tanaman) error {
    if len(users) < 3 || len(tanamans) == 0 {
        return nil
    }

    records := []models.Booking{
        {UserID: users[2].ID, TanamanID: tanamans[0].ID},
        {UserID: users[2].ID, TanamanID: tanamans[1].ID},
    }

    for i := range records {
        if err := firstOrCreate(db, &records[i],
            "user_id = ? AND tanaman_id = ?",
            records[i].UserID, records[i].TanamanID,
        ); err != nil {
            return fmt.Errorf("seed booking %d: %w", i, err)
        }
    }
    return nil
}

// --- util ---

func firstOrCreate[T any](db *gorm.DB, dest *T, query string, args ...interface{}) error {
    var existing T
    err := db.Where(query, args...).First(&existing).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return db.Create(dest).Error
    }
    if err != nil {
        return err
    }
    *dest = existing
    return nil
}

func mustParseDate(value string) time.Time {
    t, err := time.Parse("2006-01-02", value)
    if err != nil {
        log.Fatalf("format tanggal salah (%s): %v", value, err)
    }
    return t
}

func mustParseDatePtr(value string) *time.Time {
    t := mustParseDate(value)
    return &t
}