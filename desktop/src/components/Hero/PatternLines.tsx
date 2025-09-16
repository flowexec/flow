
export function PatternLines({ size = 200, className = "", ...others }) {
  return (
    <svg
      aria-hidden
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      stroke="currentColor"
      strokeWidth="1"
      opacity="0.15"
      viewBox="0 0 200 200"
      width={size}
      height={size}
      className={className}
      {...others}
    >
    <defs>
        <style>
            {`
    .flow-line {
      stroke-dasharray: 3,2;
      animation: flow 3s linear infinite;
    }
    .flow-line-alt {
      stroke-dasharray: 4,3;
      animation: flow 4s linear infinite;
    }
    @keyframes flow {
      to {
        stroke-dashoffset: -20;
      }
    }
  `}
        </style>
    </defs>

    {/* Horizontal flowing lines */}
    <path d="M0,40 Q50,30 100,40 T200,40" className="flow-line" />
    <path d="M0,80 Q50,70 100,80 T200,80" className="flow-line-alt" />
    <path d="M0,120 Q50,110 100,120 T200,120" className="flow-line" />
    <path d="M0,160 Q50,150 100,160 T200,160" className="flow-line-alt" />

    {/* Vertical connecting lines */}
    <path d="M40,0 Q30,50 40,100 T40,200" strokeDasharray="2,3" />
    <path d="M100,0 Q90,50 100,100 T100,200" strokeDasharray="3,2" />
    <path d="M160,0 Q150,50 160,100 T160,200" strokeDasharray="2,3" />

    {/* Connection nodes */}
    <circle cx="40" cy="40" r="2" fill="currentColor" opacity="0.6" />
    <circle cx="100" cy="80" r="2" fill="currentColor" opacity="0.6" />
    <circle cx="160" cy="120" r="2" fill="currentColor" opacity="0.6" />
    <circle cx="40" cy="160" r="2" fill="currentColor" opacity="0.6" />
  </svg>
  );
}
