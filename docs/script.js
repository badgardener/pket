const REPO = "https://github.com/badgardener/pket";
const API = "https://api.github.com/repos/badgardener/pket/releases";

const docs = {
  "getting-started": [
    "Getting started",
    "START HERE",
    "Build, install, and manage local packages with a small command surface.",
    `
      <h2 id="overview">A small tool with a clear contract</h2>
      <p>
        <code>pket</code> is a package builder and package manager for applications and command-line tools.
        It creates compressed <code>.pkt</code> archives, verifies package contents with SHA-512,
        installs packages transactionally, and manages executable links.
      </p>

      <h3 id="install">Install pket</h3>
      <p>Download a binary or build from source with Go 1.27.1 or newer.</p>
      ${code(`git clone ${REPO}\ncd pket\ngo build -o pket .\n./pket --version`)}

      <h3 id="workflow">The basic workflow</h3>
      ${code(`pket build ./my-app
    pket install ./my-app-1.0.0.pkt
    pket list
    pket info my-app
    pket uninstall my-app`)}
    `,
  ],

  installation: [
    "Installation",
    "GETTING STARTED",
    "Build from source or choose a prebuilt cross-platform binary.",
    `
      <h2 id="source">From source</h2>
      <p>
        pket requires Go 1.27.1 or newer. The repository includes a build script for cross-compiling the CLI.
      </p>
      ${code(`git clone ${REPO}\ncd pket\ngo build -o pket .`)}

      <h3 id="binary">Use the executable</h3>
      <p>Run the executable directly or place it somewhere on your <code>PATH</code>.</p>
      ${code("./pket --version")}

      <h3 id="platforms">Supported targets</h3>
      <p>
        Release builds cover Linux, Windows, macOS, FreeBSD, OpenBSD, NetBSD, DragonFly BSD,
        illumos, Solaris, and Plan 9.
      </p>
    `,
  ],

  cli: [
    "CLI reference",
    "REFERENCE",
    "The complete command surface, kept intentionally small.",
    `
      <h2 id="usage">Usage</h2>
      ${code("pket [--verbose] <command> [arguments]")}

      <h3 id="commands">Commands</h3>
      ${code(`pket build <directory>      Build a package from a directory
    pket install <package>      Install or update a package archive
    pket list                   List installed packages
    pket info <package>         Show package information
    pket uninstall <package>    Uninstall by name or UID
    pket --help                 Show help
    pket --version              Show version`)}

      <h3 id="verbose">Verbose output</h3>
      <p>
        Use <code>--verbose</code> before commands that support detailed callback output.
        It is not supported by <code>list</code> or <code>info</code>.
      </p>
      ${code(`pket --verbose build ./my-app
    pket --verbose install ./my-app-1.0.0.pkt
    pket --verbose uninstall my-app`)}
    `,
  ],

  "building-packages": [
    "Building packages",
    "PACKAGES",
    "A project directory plus one manifest produces a portable archive.",
    `
      <h2 id="manifest">Create a manifest</h2>
      <p>
        The project root must contain <code>pket-package.toml</code>. The manifest names the package,
        selects payload files, and declares at least one executable.
      </p>
      ${code(`[package]
    name = "My App"
    pack = "my-app"
    version = "1.0.0"
    description = "An example packaged application."
    authors = ["Example Author"]

    [files]
    base = "build"
    assets = ["LICENSE", "README.md"]

    [[executable]]
    path = "my-app"
    link = "my-app"

    [install]
    postinstall = "pket --version"`)}

      <h3 id="build">Build the archive</h3>
      ${code("pket build ./my-app")}
      <p>The resulting archive is written inside the source directory as <code>&lt;package.pack&gt;-&lt;package.version&gt;.pkt</code>.</p>
    `,
  ],

  "package-format": [
    ".pkt format",
    "PACKAGES",
    "A .pkt file is a gzip-compressed tar archive with a predictable internal shape.",
    `
      <h2 id="archive">Inside the archive</h2>
      <div class="file-tree">
        <div class="tree-dir">my-app-1.0.0.pkt</div>
        <div>├── <span class="tree-dir">payload/</span> <span class="tree-detail">application files</span>
</div>
        <div>└── <span class="tree-dir">pket-manifest/</span>
</div>
        <div>　　├── <span class="tree-file">pket-config.toml</span> <span class="tree-detail">metadata</span>
</div>
        <div>　　└── <span class="tree-file">sha-512.sums</span> <span class="tree-detail">integrity hash</span>
</div>
      </div>

      <h3 id="verification">Verification</h3>
      <p>
        During installation, pket extracts to a temporary directory, checks archive paths, verifies the
        payload against the SHA-512 manifest when present, and only then stages and activates the package.
        Archives without sums require explicit confirmation.
      </p>

      <h3 id="layout">Installed layout</h3>
      ${code(`<package-uid>/
    ├── payload/
    ├── pket-manifest/
    │   ├── pket-config.toml
    │   ├── sha-512.sums
    │   └── files.lst
    └── <package-name>.pkt`)}
    `,
  ],

  "installing-packages": [
    "Installing packages",
    "PACKAGES",
    "Installation is staged, checked, and activated as one package operation.",
    `
      <h2 id="install">Install or update</h2>
      ${code("pket install ./my-app-1.0.0.pkt")}
      <ol>
        <li>The archive is extracted into a temporary directory.</li>
        <li>Archive paths and links are checked to prevent extraction outside that directory.</li>
        <li>Payload files are verified against the package SHA-512 manifest when present.</li>
        <li>The package is prepared in a staging directory.</li>
        <li>The staging directory is atomically activated as the installed package.</li>
        <li>Executable links are created in the user executable directory.</li>
      </ol>

      <h3 id="locations">Package data</h3>
      <p>
        Unix systems use <code>~/.local/share/pket</code>, macOS uses
        <code>~/Library/Application Support/pket</code>, and Windows uses
        <code>~/AppData/Roaming/pket</code>.
      </p>
    `,
  ],

  configuration: [
    "Configuration",
    "REFERENCE",
    "Manifest fields are explicit and map directly to package metadata.",
    `
      <h2 id="fields">Manifest fields</h2>
      ${code(`[package]   name, pack, version, description, authors
    [files]     base, assets
    [[executable]] path, link
    [install]   preinstall, postinstall`)}

      <h3 id="package">Package and files</h3>
      <p>
        <code>name</code> is human-readable. <code>pack</code> is the package UID and archive name component.
        <code>version</code> has three numeric parts. <code>files.base</code> becomes the payload source and
        <code>files.assets</code> are copied into <code>payload/pket-assets</code>.
      </p>

      <h3 id="hooks">Install commands</h3>
      <p>
        <code>preinstall</code> and <code>postinstall</code> are optional. They run through the platform shell
        with the installation directory as the working directory and require two confirmations.
        Use <code>@res</code> for the package payload directory.
      </p>
    `,
  ],

  troubleshooting: [
    "Troubleshooting",
    "REFERENCE",
    "Common failures and the checks that resolve them.",
    `
      <h2 id="checks">Common checks</h2>

      <h3>Manifest missing</h3>
      <p>Add <code>pket-package.toml</code> to the project root before building.</p>

      <h3>Executable link cannot be created</h3>
      <p>Make sure the requested link name is not already used and that the selected executable directory is writable.</p>

      <h3>Package update is declined</h3>
      <p>Run the install command again and confirm the update prompt. Existing installations are never replaced without confirmation.</p>

      <h3>SHA-512 verification fails</h3>
      <p>Rebuild the package from the intended source directory. A modified archive or payload will not pass verification.</p>
    `,
  ],
};

const order = Object.keys(docs);

function esc(s) {
  return s.replace(
    /[&<>]/g,
    (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" })[c],
  );
}

function code(s) {
  return `<div class="code-block">
<button class="copy-button" data-copy="${encodeURIComponent(s)}">copy</button>
<pre>
<code>${esc(s)}</code>
</pre>
</div>`;
}

function home() {
  return `<div class="container">
<section class="hero">
<div>
<div class="eyebrow">local package manager / go</div>
<h1>Keep your packages <span class="accent">close.</span>
</h1>
<p class="hero-copy">pket builds, verifies, installs, and manages local packages for applications and command-line tools. One small binary. One clear archive format.</p>
<div class="hero-actions">
<a class="button" href="#/downloads">Download pket <span>→</span>
</a>
<a class="button secondary" href="#/docs/getting-started">Read the docs</a>
</div>
<div class="hero-meta">
<span>license <strong>MIT</strong>
</span>
<a href="${REPO}">source ↗</a>
</div>
</div>
<div class="terminal">
<div class="terminal-bar">
<i class="dot">
</i>
<i class="dot">
</i>
<i class="dot">
</i>
<span class="terminal-title">~/projects/my-app</span>
</div>
<pre>
<span class="comment"># build a package</span>\n<span class="prompt">$</span> <span class="command">pket build .</span>\n<span class="output">reading pket-package.toml...</span>\n<span class="success">package built: my-app-1.0.0.pkt</span>\n\n<span class="comment"># install it anywhere</span>\n<span class="prompt">$</span> <span class="command">pket install my-app-1.0.0.pkt</span>\n<span class="success">SHA sums matched.</span>\n<span class="success">package activated.</span>\n\n<span class="prompt">$</span> <span class="command">pket list</span>\n<span class="output">my-app    1.0.0    installed</span>
</pre>
</div>
</section>
<section class="section">
<div class="section-heading">
<div>
<div class="eyebrow">the format</div>
<h2>Readable by design.</h2>
</div>
<p class="section-intro">A .pkt archive is just a compressed tar with a small, inspectable layout.</p>
</div>
<div class="format-band">
<div class="format-copy">
<p>Package metadata lives beside the payload. A SHA-512 manifest lets the installer verify the files before activation. No registry is required.</p>
<a class="button secondary" href="#/docs/package-format">Explore the .pkt format →</a>
</div>
<div class="file-tree">
<div class="tree-dir">my-app-1.0.0.pkt</div>
<div>├── <span class="tree-dir">payload/</span>
</div>
<div>└── <span class="tree-dir">pket-manifest/</span>
</div>
<div>　　├── <span class="tree-file">pket-config.toml</span>
</div>
<div>　　└── <span class="tree-file">sha-512.sums</span>
</div>
</div>
</div>
</section>
<section class="section">
<div class="section-heading">
<div>
<div class="eyebrow">what it does</div>
<h2>Small surface. Useful edges.</h2>
</div>
<p class="section-intro">The pieces you need for local distribution, without a service in the middle.</p>
</div>
<div class="feature-grid">${[
    ["01", "Local packages", "Build and install from paths on disk."],
    ["02", "Fast builds", "Parallel gzip compression and decompression."],
    ["03", "SHA-512 checks", "Verify the payload before activation."],
    ["04", "Cross-platform", "Release binaries for a wide range of targets."],
    ["05", "Executables", "Create links in a suitable writable bin directory."],
    ["06", "Simple config", "One TOML manifest describes a package."],
    ["07", "Post-install", "Optional hooks with explicit confirmation."],
    ["08", "Transactional", "Stage and activate packages as a unit."],
  ]
    .map(
      (x) =>
        `<div class="feature">
<span class="feature-index">${x[0]}</span>
<h3>${x[1]}</h3>
<p>${x[2]}</p>
</div>`,
    )
    .join("")}</div>
</section>
<section class="section">
<div class="why-grid">
<div class="why-copy">
<div class="eyebrow">why pket?</div>
<h2>Fewer moving parts.</h2>
<p>pket is for the moments when a local package archive and a dependable install path are enough. The format is plain, the commands are finite, and the installed state stays on your machine.</p>
</div>
<div class="facts">
<div class="fact">
<span>archive</span>
<strong>gzip + tar</strong>
</div>
<div class="fact">
<span>manifest</span>
<strong>TOML</strong>
</div>
<div class="fact">
<span>integrity</span>
<strong>SHA-512</strong>
</div>
<div class="fact">
<span>source</span>
<strong>Go 1.27.1+</strong>
</div>
</div>
</div>
</section>
</div>`;
}

function docsPage(slug) {
  const d = docs[slug] || docs[order[0]];
  return `<div class="container">
<div class="page-head">
<div class="eyebrow">${d[1]}</div>
<h1>${d[0]}</h1>
<p>${d[2]}</p>
</div>
<div class="docs-layout">
<aside class="docs-sidebar">
<h4>Documentation</h4>
<nav>${order.map((k) => `<a class="${k === slug ? "active" : ""}" href="#/docs/${k}">${docs[k][0]}</a>`).join("")}</nav>
</aside>
<article class="docs-content">${d[3]}<div class="doc-nav">${
    order[order.indexOf(slug) - 1]
      ? `<a href="#/docs/${order[order.indexOf(slug) - 1]}">
<span>← Previous</span>${docs[order[order.indexOf(slug) - 1]][0]}</a>`
      : ""
  }${
    order[order.indexOf(slug) + 1]
      ? `<a href="#/docs/${order[order.indexOf(slug) + 1]}">
<span>Next</span>${docs[order[order.indexOf(slug) + 1]][0]} →</a>`
      : ""
  }</div>
</article>
<aside class="toc">
<a href="#overview">On this page</a>
<a href="#install">Install</a>
<a href="#workflow">Workflow</a>
</aside>
</div>
</div>`;
}

function downloads() {
  return `<div class="container">
<div class="page-head download-head">
<div>
<div class="eyebrow">binaries</div>
<h1>Download pket.</h1>
<p>Pick a release asset for your operating system and architecture. URLs come directly from GitHub Releases.</p>
</div>
<div class="release-status loading" id="release-status">loading releases...</div>
</div>
<div id="downloads-content">
<div class="empty">Fetching public release data from api.github.com...</div>
</div>
</div>`;
}

function date(s) {
  return new Intl.DateTimeFormat("en", {
    year: "numeric",
    month: "short",
    day: "numeric",
  }).format(new Date(s));
}

function observeReveals(root = document) {
  const elements = root.querySelectorAll(
    ".section, .page-head, .docs-sidebar, .docs-content > *, .toc, .release-panel, .release-list, .asset-group, .release-item, .doc-nav",
  );

  elements.forEach((element, index) => {
    element.classList.add("reveal");
    if (index % 4) element.classList.add(`reveal-delay-${index % 4}`);
  });

  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    elements.forEach((element) => element.classList.add("is-visible"));
    return;
  }

  const observer = new IntersectionObserver(
    (entries, currentObserver) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        entry.target.classList.add("is-visible");
        currentObserver.unobserve(entry.target);
      });
    },
    { threshold: 0.12, rootMargin: "0px 0px -8%" },
  );

  elements.forEach((element) => observer.observe(element));
}

function randomizeAmbient() {
  const randomBetween = (minimum, maximum) =>
    minimum + Math.random() * (maximum - minimum);
  const scene = document.querySelector(".ambient-scene");
  const pageWidth = Math.max(scene.clientWidth, 1);
  const pageHeight = Math.max(
    document.documentElement.scrollHeight,
    document.body.scrollHeight,
    window.innerHeight,
  );

  document.querySelectorAll(".ambient-field").forEach((field) => {
    const fieldWidth = field.offsetWidth;
    const fieldHeight = field.offsetHeight;
    const horizontalSpace = Math.max(pageWidth - fieldWidth, 0);
    const verticalSpace = Math.max(pageHeight - fieldHeight, 0);
    const startX = randomBetween(0, horizontalSpace);
    const startY = randomBetween(0, verticalSpace);
    const endX = randomBetween(0, horizontalSpace);
    const endY = randomBetween(0, verticalSpace);

    field.style.left = `${startX.toFixed(1)}px`;
    field.style.top = `${startY.toFixed(1)}px`;
    field.style.right = "auto";
    field.style.bottom = "auto";
    field.style.setProperty("--ambient-x", `${(endX - startX).toFixed(1)}px`);
    field.style.setProperty("--ambient-y", `${(endY - startY).toFixed(1)}px`);
    field.style.setProperty(
      "--ambient-opacity",
      randomBetween(0.72, 0.92).toFixed(2),
    );
    field.style.animationDuration = `${randomBetween(22, 42).toFixed(2)}s`;
    field.style.animationDelay = `-${randomBetween(0, 18).toFixed(2)}s`;
  });
}

function os(a) {
  const n = a.name.toLowerCase();
  return (
    [
      ["dragonfly", "DragonFly BSD"],
      ["freebsd", "FreeBSD"],
      ["openbsd", "OpenBSD"],
      ["netbsd", "NetBSD"],
      ["illumos", "illumos"],
      ["solaris", "Solaris"],
      ["plan9", "Plan 9"],
      ["darwin", "macOS"],
      ["windows", "Windows"],
      ["linux", "Linux"],
    ].find((x) => n.includes(`-${x[0]}-`))?.[1] || "Other"
  );
}

function arch(a) {
  const n = a.name.toLowerCase();
  return (
    [
      "amd64",
      "arm64",
      "arm",
      "386",
      "riscv64",
      "ppc64le",
      "ppc64",
      "mips64le",
      "mips64",
      "mipsle",
      "mips",
      "loong64",
      "s390x",
    ].find((x) => n.includes(`-${x}`)) || "binary"
  );
}

async function load() {
  const st = document.querySelector("#release-status"),
    out = document.querySelector("#downloads-content");

  try {
    const r = await fetch(API, {
      headers: { Accept: "application/vnd.github+json" },
    });

    if (!r.ok) throw Error();
    const releases = await r.json();
    if (!releases.length) throw Error();
    const latest = releases[0],
      groups = {};
    (latest.assets || []).forEach((a) => (groups[os(a)] ??= []).push(a));
    st.textContent = `${releases.length} release${releases.length === 1 ? "" : "s"} available`;
    st.className = "release-status";
    out.innerHTML = `<section class="release-panel">
<div class="release-top">
<h2>Latest release <span class="accent">${latest.tag_name}</span>
</h2>
<div>
<time>${date(latest.published_at || latest.created_at)}</time> · <a href="${latest.html_url}">release notes ↗</a>
</div>
</div>
<div class="asset-groups">${
      Object.entries(groups)
        .map(
          ([name, items]) =>
            `<section class="asset-group">
<h3>${name}<span class="asset-count">${items.length} assets</span>
</h3>${items
              .map(
                (a) => `<div class="asset-row">
<span>${arch(a)} · ${a.name}</span>
<a href="${a.browser_download_url}" download>download ↓</a>
</div>`,
              )
              .join("")}</section>`,
        )
        .join("") ||
      "<div class=empty>This release has no downloadable assets.</div>"
    }</div>
</section>
<h2>All releases</h2>
<div class="release-list">${releases
      .map(
        (x) => `<a class="release-item" href="${x.html_url}">
<strong>${x.tag_name}</strong>
<time>${date(x.published_at || x.created_at)} · ${x.assets?.length || 0} assets ↗</time>
</a>`,
      )
      .join("")}</div>`;
    document.querySelector("#nav-version")?.replaceChildren(latest.tag_name);
    observeReveals(out);
    requestAnimationFrame(randomizeAmbient);
  } catch (e) {
    st.textContent = "could not load releases";
    st.className = "release-status error";
    out.innerHTML = `<div class="empty">GitHub Releases could not be loaded right now. Check your connection or visit <a href="${REPO}/releases">the releases page ↗</a> directly.</div>`;
  }
}

function render() {
  const bits = location.hash.replace(/^#\/?/, "").split("/"),
    route = bits[0],
    root = document.querySelector("#app");
  document
    .querySelectorAll("[data-route-link]")
    .forEach((x) =>
      x.classList.toggle(
        "active",
        x.dataset.routeLink === (route === "docs" ? "docs" : route || "home"),
      ),
    );
  if (route === "docs") root.innerHTML = docsPage(bits[1] || "getting-started");
  else if (route === "downloads") {
    root.innerHTML = downloads();
    observeReveals(root);
    load();
  } else {
    root.innerHTML = home();
    observeReveals(root);
  }
  requestAnimationFrame(randomizeAmbient);
  window.scrollTo(0, 0);
  document.querySelector(".main-nav").classList.remove("open");
}

document.addEventListener("click", (e) => {
  const b = e.target.closest(".copy-button");
  if (b) {
    navigator.clipboard?.writeText(decodeURIComponent(b.dataset.copy));
    b.textContent = "copied";
    setTimeout(() => (b.textContent = "copy"), 1200);
  }
});

document.querySelector("#menu-button").addEventListener("click", (e) => {
  const n = document.querySelector(".main-nav"),
    open = n.classList.toggle("open");
  e.currentTarget.setAttribute("aria-expanded", open);
  e.currentTarget.setAttribute(
    "aria-label",
    open ? "Close navigation" : "Open navigation",
  );
});

window.addEventListener("hashchange", render);
window.addEventListener("resize", () =>
  requestAnimationFrame(randomizeAmbient),
);
render();
