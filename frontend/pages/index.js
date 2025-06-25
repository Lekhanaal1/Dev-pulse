export default function Home() {
  return (
    <main style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginTop: 80 }}>
      <h1>🚀 DevPulse</h1>
      <p>Developer Task Tracker — Microservices Demo</p>
      <ul style={{ textAlign: 'left', marginTop: 32 }}>
        <li>Go (Gin) microservices</li>
        <li>NATS async messaging</li>
        <li>PostgreSQL & MongoDB</li>
        <li>Prometheus + Grafana monitoring</li>
        <li>Docker Compose orchestration</li>
      </ul>
    </main>
  );
} 