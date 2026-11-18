<?php
array_filter([0]);
array_filter([false, 2]);
array_filter([0, 2], fn($v) => true);
array_filter($unknown);
array_filter([0, false]);
array_filter([0, "a"], callback: fn($v) => $v !== null);
array_filter(...$args);
other([0,1]);
