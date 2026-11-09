-- Inicialização do banco de dados para verificação de persistência
CREATE TABLE IF NOT EXISTS system_telemetry (
    id SERIAL PRIMARY KEY,
    component VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO system_telemetry (component, status) VALUES 
('nginx-proxy', 'initialized'),
('api-node-1', 'ready'),
('api-node-2', 'ready'),
('redis-cache', 'connected');
