FROM globalstudent/wbfy-base:latest

USER root

# Install Go
RUN apt-get update && apt-get install -y \
    golang \
    && rm -rf /var/lib/apt/lists/*

# Set Go environment variables
ENV GOPATH=/home/student/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Set default command
ENV WBFY_CMD="go run"

# Switch back to student user
USER student

# Create Go workspace
RUN mkdir -p $GOPATH/{src,pkg,bin}

# Add Go env to .bashrc
RUN echo 'export GOPATH=$HOME/go' >> ~/.bashrc && \
    echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~/.bashrc
