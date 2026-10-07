<?php
namespace Blog;

use function Text\Slugify as <weak_warning descr="Alias slugify repeats the imported name; remove it.">slugify</weak_warning>;
use function Text\{Excerpt as <weak_warning descr="Alias excerpt repeats the imported name; remove it.">excerpt</weak_warning>, wordCount as count_words};
use Text\{function Title as <weak_warning descr="Alias TITLE repeats the imported name; remove it.">TITLE</weak_warning>, Renderer as renderer};
use const Text\MAX_LEN as max_len;
use Text\Formatter as formatter;
