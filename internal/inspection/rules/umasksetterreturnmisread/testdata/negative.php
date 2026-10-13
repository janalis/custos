<?php
umask(0077); if (umask() !== 0077) { throw new RuntimeException(); }
