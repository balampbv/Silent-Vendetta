import React from 'react';

interface MaskProps {
  width?: number;
  height?: number;
  className?: string;
}

const VendettaMask: React.FC<MaskProps> = ({ width = 32, height = 32, className = '' }) => {
  return (
    <svg 
      width={width} 
      height={height} 
      viewBox="0 0 32 32" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {/* Mask outline */}
      <path d="M8 10C8 10 10 8 16 8C22 8 24 10 24 10L25 12C25 12 26 14 26 16C26 20 22 24 16 24C10 24 6 20 6 16C6 14 7 12 7 12L8 10Z" fill="currentColor"/>
      
      {/* Eyebrows */}
      <path d="M11 13C11 13 13 12 16 12C19 12 21 13 21 13" stroke="#1a1a1a" strokeWidth="1.5" strokeLinecap="round"/>
      <path d="M11 14L13 15" stroke="#1a1a1a" strokeWidth="1.5" strokeLinecap="round"/>
      <path d="M21 14L19 15" stroke="#1a1a1a" strokeWidth="1.5" strokeLinecap="round"/>
      
      {/* Eyes */}
      <path d="M12 16C12 16 13 17 14 17C15 17 16 16 16 16" fill="#1a1a1a"/>
      <path d="M16 16C16 16 17 17 18 17C19 17 20 16 20 16" fill="#1a1a1a"/>
      
      {/* Mustache and Smile */}
      <path d="M13 19C13 19 14 20 16 20C18 20 19 19 19 19" stroke="#1a1a1a" strokeWidth="1.5" strokeLinecap="round"/>
      <path d="M16 19L16 21" stroke="#1a1a1a" strokeWidth="1.5" strokeLinecap="round"/>
    </svg>
  );
};

export default VendettaMask; 