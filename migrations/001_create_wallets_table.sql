CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,             
    balance FLOAT NOT NULL DEFAULT 0.0, 
    created_at TIMESTAMP DEFAULT now(),       
    updated_at TIMESTAMP DEFAULT now()        
);
