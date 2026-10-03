import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router'
import warchestIcon from './warchestemoticon.png';

interface Warchest {
  id: string;
  x: number;
  y: number;
  orientation: number;
  size: number;
}

export default function LandingPage() {
  const [text, setText] = useState("");
  const [warchests, setWarchests] = useState<Warchest[]>([]);
  const navigate = useNavigate();

  function handleChange(change: string) {
    if (change.length > text.length) {
      for (var i = 0; i < (change.length - text.length); i++) {
        const newWarchest: Warchest = {
          id: crypto.randomUUID(),
          x: Math.random() * 100,
          y: Math.random() * 100,
          orientation: Math.random() * 360,
          size: Math.random() * 20,
        };
        setWarchests((prev) => [...prev, newWarchest])
      }
    } else if (change.length < text.length) {
      setWarchests((current) => current.slice(0, -(text.length - change.length)));
    }
  }

  function handleInput(change: string) {
    handleChange(change);
    setText(change);
  }

  function handleSubmit() {
    if (text.toLowerCase() === "warchest") {
      navigate({ to: '/dashboard' })
      return;
    }
  }

  return (
    <div>
      <div
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          zIndex: -1,
          pointerEvents: 'none',
        }}
      >
        {warchests.map((warchest) => (
          <img
            key={warchest.id}
            src={warchestIcon}
            alt="Warchest"
            style={{
              position: 'absolute',
              left: `${warchest.x}vw`,
              top: `${warchest.y}vh`,
              width: `${warchest.size}vw`,
              height: 'auto',
              transform: `rotate(${warchest.orientation}deg)`,
              pointerEvents: 'none',
              userSelect: 'none',
            }}
          />
        ))}
      </div>
      <div>

      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          flexDirection: 'column', 
          alignItems: 'center',
          height: '100vh',
        }}
      >
        <h1>WARCHEST</h1>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            value={text}
            onChange={(e) => handleInput(e.target.value)}
            placeholder="Name the best port here..."
            style={{
              padding: '15px',
              fontSize: '18px',
              borderRadius: '8px',
              border: '1px solid #ccc',
              background: 'rgba(0,0,0,0)',
              width: '300px',
              boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
            }}
          />
        </form>
      </div>
    </div>
  );
}