FROM php:8.2-apache

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       mariadb-server mariadb-client default-libmysqlclient-dev libzip-dev unzip \
       libpng-dev libjpeg-dev libfreetype6-dev procps \
    && docker-php-ext-configure gd --with-freetype --with-jpeg \
    && docker-php-ext-install -j$(nproc) gd pdo_mysql mysqli \
    && a2enmod rewrite \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /var/www/html

COPY docker-php-ext-gd.ini /usr/local/etc/php/conf.d/docker-php-ext-gd.ini
COPY docker-php-ext-pdo_mysql.ini /usr/local/etc/php/conf.d/docker-php-ext-pdo_mysql.ini
COPY docker-php-ext-mysqli.ini /usr/local/etc/php/conf.d/docker-php-ext-mysqli.ini

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 80 3306

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["apache2-foreground"]
