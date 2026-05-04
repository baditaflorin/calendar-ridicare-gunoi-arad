# ADR 0008: Responsive Mobile-First Overhaul & Premium Branding

## Status
Accepted

## Context
The "Reciclare Arad" platform has evolved from a simple utility to a premium civic service. However, the initial layout was desktop-centric, leading to a poor user experience on mobile devices. Additionally, the branding needed to be more engaging ("cute") to encourage family participation and children's involvement in recycling.

## Decision
We will implement a comprehensive responsive overhaul of the entire web interface using a mobile-first approach. Key changes include:
1.  **Mobile-First Layout**: Use flexible grids and flexbox to ensure the interface scales seamlessly from 320px to 1440px.
2.  **Typography**: Adopt rounded, friendly fonts (Fredoka and Quicksand) to enhance readability and "cute" appeal.
3.  **Component Refactoring**: 
    - The search bar will stack vertically on mobile.
    - The calendar grid will use a more compact view or scroll horizontally on very small screens.
    - Modal windows will use `max-height` and scrolling to prevent overflow.
4.  **Premium Aesthetics**: Introduce vibrant but soft pastel colors, subtle micro-animations, and mascot-based storytelling.

## Consequences
- **Positive**: Significantly improved user retention on mobile devices. Increased engagement from families and children. Professional, "premium" feel that justifies the print ordering service.
- **Negative**: Increased complexity in the CSS file. Need to carefully test layout changes across different device resolutions.
