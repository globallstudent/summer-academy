FROM globalstudent/wbfy-base:latest

USER root

# Install Node.js and npm
RUN apt-get update && apt-get install -y \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

# Install nvm for version management
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# Install common Node.js packages globally
RUN npm install -g \
    nodemon \
    jest \
    express-generator

# Set default command
ENV WBFY_CMD="node"

# Switch back to student user
USER student

# Configure nvm for student user
RUN echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.bashrc && \
    echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> ~/.bashrc
