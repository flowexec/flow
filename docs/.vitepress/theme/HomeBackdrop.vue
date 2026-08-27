<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

/**
 * Ambient backdrop for the home page.
 *
 * A drifting node field — the same family as the one on jahvon.dev — but the
 * edges here carry pulses that travel from node to node and pop when they
 * land. It reads as work moving through a graph, which is what flow does.
 *
 * Vanilla canvas, no library. Colours are read from the theme's CSS variables
 * so it follows dark/light without a second palette, and it holds a single
 * static frame when the visitor asks for reduced motion.
 */

const canvas = ref<HTMLCanvasElement | null>(null)

interface Node {
  x: number
  y: number
  vx: number
  vy: number
  r: number
}

interface Pulse {
  from: number
  to: number
  /** 0 → 1 along the edge, then a short pop while > 1. */
  t: number
  speed: number
}

const LINK_DISTANCE = 150
const MAX_PULSES = 5

onMounted(() => {
  const el = canvas.value
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return

  const reduced = matchMedia('(prefers-reduced-motion: reduce)').matches
  const dpr = Math.min(window.devicePixelRatio || 1, 2)

  let width = 0
  let height = 0
  let nodes: Node[] = []
  let pulses: Pulse[] = []
  let frame: number | null = null
  let running = false
  let phase = 0

  let cEdge = ''
  let cNode = ''
  let cPulse = ''

  function readColors() {
    const style = getComputedStyle(document.documentElement)
    cEdge = style.getPropertyValue('--vp-c-divider').trim() || '#ddd'
    cNode = style.getPropertyValue('--vp-c-text-3').trim() || '#999'
    cPulse = style.getPropertyValue('--vp-c-brand-1').trim() || '#7FBBB3'
  }

  function resize() {
    width = el!.clientWidth
    height = el!.clientHeight
    el!.width = width * dpr
    el!.height = height * dpr
    ctx!.setTransform(dpr, 0, 0, dpr, 0, 0)
  }

  function seed() {
    const count = Math.max(10, Math.min(28, Math.round((width * height) / 46000)))
    nodes = Array.from({ length: count }, () => ({
      x: Math.random() * width,
      y: Math.random() * height,
      vx: (Math.random() - 0.5) * 0.13,
      vy: (Math.random() - 0.5) * 0.13,
      r: 1.3 + Math.random() * 1.3,
    }))
    pulses = []
  }

  function distance(a: Node, b: Node) {
    return Math.hypot(a.x - b.x, a.y - b.y)
  }

  /** Send a pulse down some edge that currently exists. */
  function spawnPulse() {
    if (nodes.length < 2) return
    const from = Math.floor(Math.random() * nodes.length)
    const candidates: number[] = []
    for (let i = 0; i < nodes.length; i++) {
      if (i !== from && distance(nodes[from], nodes[i]) < LINK_DISTANCE) candidates.push(i)
    }
    if (!candidates.length) return
    pulses.push({
      from,
      to: candidates[Math.floor(Math.random() * candidates.length)],
      t: 0,
      speed: 0.0035 + Math.random() * 0.004,
    })
  }

  function draw() {
    ctx!.clearRect(0, 0, width, height)

    // Edges
    ctx!.strokeStyle = cEdge
    ctx!.lineWidth = 1
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const d = distance(nodes[i], nodes[j])
        if (d >= LINK_DISTANCE) continue
        ctx!.globalAlpha = (1 - d / LINK_DISTANCE) * 0.45
        ctx!.beginPath()
        ctx!.moveTo(nodes[i].x, nodes[i].y)
        ctx!.lineTo(nodes[j].x, nodes[j].y)
        ctx!.stroke()
      }
    }

    // Nodes
    ctx!.fillStyle = cNode
    for (const n of nodes) {
      ctx!.globalAlpha = 0.55
      ctx!.beginPath()
      ctx!.arc(n.x, n.y, n.r, 0, Math.PI * 2)
      ctx!.fill()
    }

    // Pulses in transit, and the pop where one lands
    for (const p of pulses) {
      const a = nodes[p.from]
      const b = nodes[p.to]
      if (!a || !b) continue

      if (p.t <= 1) {
        const x = a.x + (b.x - a.x) * p.t
        const y = a.y + (b.y - a.y) * p.t
        // Fade in and out so pulses never blink on at full strength.
        ctx!.globalAlpha = Math.sin(p.t * Math.PI) * 0.85
        ctx!.fillStyle = cPulse
        ctx!.beginPath()
        ctx!.arc(x, y, 2.1, 0, Math.PI * 2)
        ctx!.fill()
      } else {
        const pop = (p.t - 1) / 0.35
        ctx!.globalAlpha = (1 - pop) * 0.5
        ctx!.strokeStyle = cPulse
        ctx!.lineWidth = 1.2
        ctx!.beginPath()
        ctx!.arc(b.x, b.y, 2 + pop * 7, 0, Math.PI * 2)
        ctx!.stroke()
      }
    }

    ctx!.globalAlpha = 1
  }

  function step() {
    for (const n of nodes) {
      n.x += n.vx + Math.sin(phase + n.y * 0.008) * 0.045
      n.y += n.vy + Math.cos(phase + n.x * 0.008) * 0.045
      if (n.x < -20) n.x = width + 20
      else if (n.x > width + 20) n.x = -20
      if (n.y < -20) n.y = height + 20
      else if (n.y > height + 20) n.y = -20
    }
    phase += 0.004

    for (const p of pulses) p.t += p.speed
    pulses = pulses.filter(
      (p) => p.t < 1.35 && distance(nodes[p.from], nodes[p.to]) < LINK_DISTANCE * 1.25
    )
    if (pulses.length < MAX_PULSES && Math.random() < 0.018) spawnPulse()
  }

  function loop() {
    if (!running) return
    step()
    draw()
    frame = requestAnimationFrame(loop)
  }

  function start() {
    if (running || reduced) return
    running = true
    loop()
  }

  function stop() {
    running = false
    if (frame !== null) cancelAnimationFrame(frame)
    frame = null
  }

  function onVisibility() {
    if (document.hidden) stop()
    else start()
  }

  let resizeTimer: number | undefined
  function onResize() {
    window.clearTimeout(resizeTimer)
    resizeTimer = window.setTimeout(() => {
      resize()
      seed()
      draw()
    }, 180)
  }

  const themeWatcher = new MutationObserver(() => {
    readColors()
    draw()
  })

  readColors()
  resize()
  seed()
  draw()
  start()

  document.addEventListener('visibilitychange', onVisibility)
  window.addEventListener('resize', onResize)
  themeWatcher.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })

  onUnmounted(() => {
    stop()
    document.removeEventListener('visibilitychange', onVisibility)
    window.removeEventListener('resize', onResize)
    themeWatcher.disconnect()
    window.clearTimeout(resizeTimer)
  })
})
</script>

<template>
  <canvas ref="canvas" class="home-backdrop" aria-hidden="true" />
</template>

<style scoped>
.home-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  display: block;
  width: 100%;
  height: 100%;
  pointer-events: none;
  opacity: 0.55;
}

@media (max-width: 700px) {
  .home-backdrop {
    opacity: 0.34;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-backdrop {
    opacity: 0.28;
  }
}
</style>
