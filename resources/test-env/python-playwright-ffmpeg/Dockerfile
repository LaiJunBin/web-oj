FROM python:3.10

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt && pip install ffmpeg-python playwright

# Install system dependencies (including ffmpeg) required by Playwright browsers
RUN apt-get update -y && \
	apt-get install -y --no-install-recommends \
		ffmpeg \
		ca-certificates \
		fonts-liberation \
		libnss3 \
		libxss1 \
		libasound2 \
		libatk1.0-0 \
		libatk-bridge2.0-0 \
		libcups2 \
		libdrm2 \
		libxrandr2 \
		libgbm1 \
		libgtk-3-0 \
		libxcomposite1 \
		libxdamage1 \
		libxext6 \
		libxfixes3 \
		libxrender1 \
		libx11-6 \
	&& rm -rf /var/lib/apt/lists/*

# Install Playwright browsers (Chromium, Firefox, WebKit) and their deps
RUN python -m playwright install --with-deps

RUN apt-get install -y mariadb-client

CMD [ "tail" , "-f", "/dev/null" ]