<?php
namespace Blog;

use function Text\Slugify;
use function Text\{Excerpt, wordCount as count_words};
use Text\{function Title, Renderer as renderer};
use const Text\MAX_LEN as max_len;
use Text\Formatter as formatter;
