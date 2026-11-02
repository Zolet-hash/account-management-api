INSERT INTO exchanges(id,name,minimum_age,maximum_transfer_amount) VALUES
('11111111-1111-4111-8111-111111111111','Atlas Exchange',18,10000.00),
('22222222-2222-4222-8222-222222222222','Aurora Exchange',21,25000.00)

ON CONFLICT(name) 
    DO UPDATE SET 
        minimum_age=EXCLUDED.minimum_age, 
        maximum_transfer_amount=EXCLUDED.maximum_transfer_amount;