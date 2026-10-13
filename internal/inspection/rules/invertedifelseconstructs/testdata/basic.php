<?php
function ready(): bool { return true; }
function maybe(): ?bool { return null; }

if (!ready()) { wait(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { go(); }

if ( ! ($left === $right) ) { differ(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { same(); }

if ($mode === 1) { one(); }
elseif (!isset($cfg['x'])) { fallback(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { useCfg(); }

if (FALSE === ready()) { wait(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { go(); }

if (false === maybe()) { a(); } else { b(); }
if (!empty($list)) { a(); } else { b(); }
if ((!ready())) { a(); } else { b(); }
if (!ready()) { a(); } else b();
if (!ready()) a(); else { b(); }
if (!$p || !$q) { a(); } else { b(); }
if (!ready()) { a(); }
