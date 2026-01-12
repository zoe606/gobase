-- Create profiles table for personalized user data
CREATE TABLE IF NOT EXISTS profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    avatar_media_id INTEGER REFERENCES media(id) ON DELETE SET NULL,
    bio TEXT,
    phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient lookups
CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_avatar ON profiles(avatar_media_id);

-- Add comment to table
COMMENT ON TABLE profiles IS 'User profile data separate from auth-related user data';
