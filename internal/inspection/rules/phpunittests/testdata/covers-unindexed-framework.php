<?php
namespace Drupal\Tests\embed\Kernel {
    use Drupal\KernelTests\KernelTestBase;

    // The framework is not indexed: the classes it covers may live there.
    class EmbedFilterTest extends KernelTestBase
    {
        /**
         * @covers \Drupal\filter\Plugin\Filter\FilterAlign
         * @covers \Drupal\filter\Plugin\Filter\FilterCaption::process
         */
        public function testFilterIntegration() {}

        /**
         * @covers ::__construct
         */
        public function <error descr="The @covers target '::__construct' cannot be resolved.">testConstruct</error>() {}
    }

    final class LocalTest
    {
        /**
         * @covers \Drupal\filter\Plugin\Filter\FilterAlign
         */
        public function <error descr="The @covers target '\Drupal\filter\Plugin\Filter\FilterAlign' cannot be resolved.">testAlign</error>() {}
    }
}
