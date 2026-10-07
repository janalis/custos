<?php
namespace Shop;

use Psr\Log\LoggerInterface as <weak_warning descr="Alias LoggerInterface repeats the imported name; remove it.">LoggerInterface</weak_warning>;
use \Redis as <weak_warning descr="Alias Redis repeats the imported name; remove it.">Redis</weak_warning>;
use Memcached as <weak_warning descr="Alias Memcached repeats the imported name; remove it.">Memcached</weak_warning>, Psr\Clock\ClockInterface as <weak_warning descr="Alias ClockInterface repeats the imported name; remove it.">ClockInterface</weak_warning>;
use function Util\slugify as <weak_warning descr="Alias slugify repeats the imported name; remove it.">slugify</weak_warning>;
use const Util\LIMIT as <weak_warning descr="Alias LIMIT repeats the imported name; remove it.">LIMIT</weak_warning>;
use Psr\Cache\{CacheItemInterface as <weak_warning descr="Alias CacheItemInterface repeats the imported name; remove it.">CacheItemInterface</weak_warning>, InvalidArgumentException};
