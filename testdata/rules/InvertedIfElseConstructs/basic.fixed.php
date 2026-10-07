<?php
function ready(): bool { return true; }
function maybe(): ?bool { return null; }

if (ready()) { go(); }
else { wait(); }

if ($left === $right) { same(); }
else { differ(); }

if ($mode === 1) { one(); }
elseif (isset($cfg['x'])) { useCfg(); }
else { fallback(); }

if (ready()) { go(); }
else { wait(); }

if (false === maybe()) { a(); } else { b(); }
if (!empty($list)) { a(); } else { b(); }
if ((!ready())) { a(); } else { b(); }
if (!ready()) { a(); } else b();
if (!ready()) a(); else { b(); }
if (!$p || !$q) { a(); } else { b(); }
if (!ready()) { a(); }
