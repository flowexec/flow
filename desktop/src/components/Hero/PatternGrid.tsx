import React from 'react';

export function PatternGrid({ size = 200, className = "", ...others }) {
    const blocks : React.ReactElement[] = [];
    const seededRandom = (seed: number) => {
        const x = Math.sin(seed) * 10000;
        return x - Math.floor(x);
    };

    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 12; col++) {
            const seed = row * 12 + col;
            blocks.push(
                <rect
                    key={`${row}-${col}`}
                    x={col * 16 + 2}
                    y={row * 24 + 4}
                    width="8"
                    height="12"
                    rx="1"
                    opacity={seededRandom(seed) > 0.7 ? 1 : 0.3}
                />
            );
        }
    }

    return (
        <svg
            aria-hidden
            xmlns="http://www.w3.org/2000/svg"
            fill="currentColor"
            opacity="0.08"
            viewBox="0 0 200 200"
            width={size}
            height={size}
            className={className}
            {...others}
        >
            {blocks}
        </svg>
    );
}
