<?php
// presets: symfony per; set: line_length=40
throw new InvalidArgumentException('A message that is long enough', 42);
throw new E(sprintf('%s', <<<EOT
    x
    EOT));
throw new E(
    // why
    'x'
);
