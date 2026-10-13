<?php
// preg_* replacements return null when PCRE fails: the cast is not redundant.
function slugify(string $title): string
{
    $slug = (string) preg_replace('/[^a-z0-9]+/', '-', strtolower($title));
    $trimmed = (string) preg_replace_callback('/-+/', fn ($m) => '-', $slug);
    $kept = (string) preg_filter('/^-|-$/', '', $trimmed);
    return <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> str_replace('--', '-', $kept);
}
