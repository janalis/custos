<?php

final class Palette
{
    final const <weak_warning descr="Declare the visibility of 'PRIMARY' explicitly.">PRIMARY</weak_warning> = '#0af', SECONDARY = '#fa0';
    final public const ACCENT = '#f0a';
    final protected const MUTED = '#999';
}

interface Themed
{
    final const <weak_warning descr="Declare the visibility of 'DEFAULT_THEME' explicitly.">DEFAULT_THEME</weak_warning> = 'light';
}
