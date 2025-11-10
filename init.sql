-- postgis for geospatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- users
CREATE TABLE IF NOT EXISTS users(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email VARCHAR(255) UNIQUE NOT NULL,
	username VARCHAR(255) UNIQUE NOT NULL,
	role VARCHAR(100) NOT NULL DEFAULT 'user',
	provider VARCHAR(100) NOT NULL,
	last_active TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- profiles
CREATE TABLE IF NOT EXISTS profiles(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id),
	fullname VARCHAR(255) NOT NULL,
	interests JSONB NOT NULL,
	bio VARCHAR(5000) NOT NULL,
	photo_url VARCHAR(500),
	photo_public_id VARCHAR(255),
	dob DATE NOT NULL CHECK(dob < NOW()),
	gender VARCHAR(10) NOT NULL,
	is_gender_public BOOLEAN NOT NULL DEFAULT true,
	relationship_intent VARCHAR(100) NOT NULL,
	-- location data stored as raw coordinates and point in postgis
	latitude DECIMAL(9, 6) NOT NULL,
	longitude DECIMAL(9, 6) NOT NULL,
	location GEOGRAPHY(POINT) NOT NULL,
	is_verified BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index for profiles location
CREATE INDEX IF NOT EXISTS idx_profiles_location ON profiles USING GIST(location);

-- swipes
CREATE TABLE IF NOT EXISTS swipes(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	swiper_id UUID NOT NULL REFERENCES users(id),
	swipee_id UUID NOT NULL REFERENCES users(id),
	swipe_type VARCHAR(10) NOT NULL CHECK(swipe_type IN ('like', 'pass')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	-- one user can swipe on the other once only
	UNIQUE(swiper_id, swipee_id)
);

-- matches
CREATE TABLE IF NOT EXISTS matches(
	-- id can possibly be a sorted concatenation of user1_id and user2_id
	id VARCHAR(255) PRIMARY KEY,
	user1_id UUID NOT NULL REFERENCES users(id),
	user2_id UUID NOT NULL REFERENCES users(id),
	-- status indicating if one user has removed match
	is_active BOOLEAN DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

-- messages
CREATE TABLE IF NOT EXISTS messages(
	id VARCHAR(255) PRIMARY KEY,
	match_id UUID NOT NULL REFERENCES matches(id),
	sender_id UUID NOT NULL REFERENCES users(id),
	content VARCHAR(5000) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);
