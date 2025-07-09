FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# Install basic tools
RUN apt-get update && apt-get install -y \
    bash \
    curl \
    git \
    nano \
    vim \
    zip \
    unzip \
    wget \
    sudo \
    procps \
    iputils-ping \
    net-tools \
    netcat \
    htop \
    less \
    man \
    tmux \
    ssh \
    jq \
    tree \
    gcc \
    g++ \
    make \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -ms /bin/bash student && \
    echo "student:student" | chpasswd && \
    adduser student sudo && \
    echo "student ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/student

# Create workspace directory
WORKDIR /workspace
RUN chown student:student /workspace

# Set working directory and switch to student user
USER student

# Add startup script
COPY --chown=student:student start.sh /start.sh
RUN chmod +x /start.sh

# Set container entrypoint
ENTRYPOINT ["/start.sh"]
