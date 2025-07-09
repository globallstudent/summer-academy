FROM globalstudent/wbfy-base:latest

USER root

# Install Python
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-dev \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Install common Python libraries
RUN pip3 install --no-cache-dir \
    numpy \
    pandas \
    matplotlib \
    requests \
    pytest \
    ipython \
    jupyter

# Set default command
ENV WBFY_CMD="python3"

# Switch back to student user
USER student

# Create a Python virtual environment in the home directory
RUN python3 -m venv ~/venv

# Add virtual environment activation to .bashrc
RUN echo 'source ~/venv/bin/activate' >> ~/.bashrc
