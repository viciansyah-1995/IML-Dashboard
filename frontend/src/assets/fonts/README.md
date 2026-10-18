# Diatype webfonts

Place the actual font files here. Expected names:

- `diatype-regular.woff2` — body text, inputs, supporting copy (400).
- `diatype-medium.woff2` — controls, navigation, labels, table headers (500).
- `diatype-bold.woff2` — headings, KPI values, emphasis (700).
- `diatype-regular-italic.woff2` — quotations and inline emphasis (400 italic).

A Diatype variable WOFF2 file with `variable` or `vf` in its name can supply normal weights instead. WOFF is supported. Use the ordinary Diatype family; Mono/Condensed/Compressed/Expanded/Extended/Rounded assets are excluded from automatic UI registration.

No font binary was available when configured. The app presently uses the system fallback unless Diatype is installed locally. Once genuine files are added, Vite registers and packages them without referencing absent URLs. Refresh the page or restart/rebuild Vite after adding files. The configuration does not fetch font files from the Dinamo specimen site.
