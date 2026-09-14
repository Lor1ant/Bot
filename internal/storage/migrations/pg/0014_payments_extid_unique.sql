CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_method_extid
    ON payments(method, ext_id)
    WHERE ext_id <> '';
