const fs = require('fs');
const path = require('path');
const { Resvg } = require('@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="amber-core" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="100%" stop-color="#d97706"/>
    </linearGradient>

    <linearGradient id="titanium-mantle" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#e2e8f0"/>
      <stop offset="100%" stop-color="#94a3b8"/>
    </linearGradient>

    <linearGradient id="obsidian-armor" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#1e293b"/>
      <stop offset="100%" stop-color="#0b0f19"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#000000" flood-opacity="0.14" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway Frame -->
      <polygon points="
        0,-410
        355,-205
        355,205
        0,410
        -355,205
        -355,-205
      " fill="none" stroke="#0f172a" stroke-width="40" stroke-linejoin="round" />

      <polygon points="
        0,-375
        324,-187
        324,187
        0,375
        -324,187
        -324,-187
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.4" stroke-dasharray="16, 12" />

      <!-- Honeycomb Defensive Sub-Grid -->
      <polygon points="
        0,-330
        285,-165
        285,165
        0,330
        -285,165
        -285,-165
      " fill="none" stroke="#f59e0b" stroke-width="2" opacity="0.25" stroke-dasharray="8, 8" />

      <!-- Watertight Solid Mascot Base Silhouette (Cyber Honey Badger / Armored Ratel) -->
      <path d="
        M 0,-280
        L 85,-270 L 160,-220 L 220,-140 L 250,-40 L 230,80 L 180,180 L 110,250 L 0,285
        L -110,250 L -180,180 L -230,80 L -250,-40 L -220,-140 L -160,-220 L -85,-270
        Z
      " fill="#090d16" />

      <!-- Mantle / Crest (Titanium White-Silver Ratel Back) -->
      <polygon points="0,-275 80,-265 140,-215 90,-160 0,-140" fill="url(#titanium-mantle)" />
      <polygon points="0,-275 -80,-265 -140,-215 -90,-160 0,-140" fill="#cbd5e1" />

      <polygon points="0,-140 90,-160 140,-215 160,-130 90,-80 0,-60" fill="#94a3b8" />
      <polygon points="0,-140 -90,-160 -140,-215 -160,-130 -90,-80 0,-60" fill="#64748b" />

      <!-- Lateral Armored Ears -->
      <polygon points="140,-215 210,-245 220,-140 160,-130" fill="#1e293b" />
      <polygon points="160,-190 200,-220 195,-155" fill="url(#cyan-glow)" />

      <polygon points="-140,-215 -210,-245 -220,-140 -160,-130" fill="#0f172a" />
      <polygon points="-160,-190 -200,-220 -195,-155" fill="url(#cyan-glow)" />

      <!-- Cheek Armor Plates (Wide Zygomatic Arches) -->
      <polygon points="160,-130 220,-140 250,-40 180,-10 90,-80" fill="#1e293b" />
      <polygon points="-160,-130 -220,-140 -250,-40 -180,-10 -90,-80" fill="#0f172a" />

      <polygon points="180,-10 250,-40 230,80 160,70 110,10" fill="#0f172a" />
      <polygon points="-180,-10 -250,-40 -230,80 -160,70 -110,10" fill="#1e293b" />

      <!-- Forehead & Central Armor Spire -->
      <polygon points="0,-60 90,-80 80,0 0,30" fill="#334155" />
      <polygon points="0,-60 -90,-80 -80,0 0,30" fill="#1e293b" />

      <!-- Amber Cryptographic Trap Core (Hexagonal Honey Node) -->
      <polygon points="
        0,-10
        40,-35
        40,-80
        0,-105
        -40,-80
        -40,-35
      " fill="url(#amber-core)" stroke="#090d16" stroke-width="4" />
      <polygon points="
        0,-25
        25,-40
        25,-70
        0,-85
        -25,-70
        -25,-40
      " fill="#fef3c7" opacity="0.8" />

      <!-- Predator Eyes (Sharp Slanted Almond Facets - Electric Cyan) -->
      <polygon points="40,-45 105,-55 75,-25 35,-30" fill="url(#cyan-glow)" stroke="#090d16" stroke-width="3" />
      <polygon points="55,-45 90,-50 70,-32" fill="#e0f2fe" />

      <polygon points="-40,-45 -105,-55 -75,-25 -35,-30" fill="url(#cyan-glow)" stroke="#090d16" stroke-width="3" />
      <polygon points="-55,-45 -90,-50 -70,-32" fill="#e0f2fe" />

      <!-- Aerodynamic Angular Muzzle (No Buck Teeth, Watertight Slate Geometry) -->
      <polygon points="0,30 80,0 70,80 0,110" fill="#1e293b" />
      <polygon points="0,30 -80,0 -70,80 0,110" fill="#0f172a" />

      <polygon points="0,110 70,80 110,10 60,140 0,180" fill="#334155" />
      <polygon points="0,110 -70,80 -110,10 -60,140 0,180" fill="#1e293b" />

      <!-- Obsidian Nose & Chin Strike Plate -->
      <polygon points="0,75 35,90 0,120 -35,90" fill="#090d16" stroke="#00f5ff" stroke-width="2" />
      <polygon points="0,180 50,150 70,210 0,260" fill="#1e293b" />
      <polygon points="0,180 -50,150 -70,210 0,260" fill="#0f172a" />

      <!-- Lower Jaw & Throat Armor -->
      <polygon points="0,260 70,210 110,250 0,285" fill="#334155" />
      <polygon points="0,260 -70,210 -110,250 0,285" fill="#1e293b" />

      <!-- Electric Cyan Accent Pips -->
      <circle cx="0" cy="150" r="5" fill="#00f5ff" />
      <circle cx="-135" cy="40" r="4" fill="#00f5ff" opacity="0.7" />
      <circle cx="135" cy="40" r="4" fill="#00f5ff" opacity="0.7" />
    </g>
  </g>
</svg>`;
}

async function main() {
  const outputDir = path.resolve(__dirname, '../docs/images');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');
  console.log('✓ Written', svgPath);

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: 1024 },
  });
  const pngData = resvg.render();
  fs.writeFileSync(pngPath, pngData.asPng());
  console.log('✓ Rendered', pngPath, '(1024x1024)');
}

main().catch(console.error);
