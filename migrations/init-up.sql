CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    is_moderator BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS drugs (
    drug_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    photo_url VARCHAR(255),
    video VARCHAR(255),
    adult_dose_mg NUMERIC(12,2) NOT NULL DEFAULT 0,
    dose_per_m2_mg NUMERIC(12,2) NOT NULL DEFAULT 0,
    max_daily_mg NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS prescriptions (
    prescription_id SERIAL PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    creator_id INTEGER NOT NULL REFERENCES users(user_id),
    forming_date TIMESTAMP,
    finish_date TIMESTAMP,
    moderator_id INTEGER REFERENCES users(user_id),
    doctor_full_name VARCHAR(200) NOT NULL,
    notes VARCHAR(2000)
);

CREATE TABLE IF NOT EXISTS prescription_drugs (
    prescription_id INTEGER NOT NULL REFERENCES prescriptions(prescription_id),
    drug_id INTEGER NOT NULL REFERENCES drugs(drug_id),
    height_cm INTEGER NOT NULL DEFAULT 0,
    weight_kg INTEGER NOT NULL DEFAULT 0,
    pediatric_dose_mg NUMERIC(12,2),
    PRIMARY KEY (prescription_id, drug_id)
);

INSERT INTO users (login, password, is_moderator) VALUES
    ('user1', 'pass1', false),
    ('moderator', 'modpass', true)
ON CONFLICT (login) DO NOTHING;

INSERT INTO drugs (title, description, is_deleted, photo_url, video, adult_dose_mg, dose_per_m2_mg, max_daily_mg) VALUES
    ('Парацетамол', 'Жаропонижающее и обезболивающее. Рекомендован при лихорадке и боли лёгкой и средней интенсивности.', false, 'paracetamol.jpg', 'paracetamol.mp4', 1000, 250, 4000),
    ('Ибупрофен', 'НПВП, жаропонижающее и противовоспалительное. Применяется при боли и воспалении.', false, 'ibuprofen.jpg', 'ibuprofen.mp4', 400, 100, 2400),
    ('Амоксициллин', 'Антибиотик группы пенициллинов. Назначается при бактериальных инфекциях дыхательных путей и ЛОР-органов.', false, 'amoxicillin.jpg', 'amoxicillin.mp4', 500, 25, 1500),
    ('Цетиризин', 'Антигистаминный препарат. Показан при аллергическом рините и крапивнице.', false, 'cetirizine.jpg', 'cetirizine.mp4', 10, 5, 10),
    ('Омепразол', 'Ингибитор протонной помпы. Используется при ГЭРБ и язвенной болезни.', false, 'omeprazole.jpg', 'omeprazole.mp4', 20, 10, 40),
    ('Домперидон', 'Противорвотное, прокинетик. При тошноте и функциональных нарушениях ЖКТ.', false, 'domperidone.jpg', 'domperidone.mp4', 10, 2.5, 30);
